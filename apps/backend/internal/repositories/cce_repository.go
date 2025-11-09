package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/query"
)

type brandProfileRepository struct {
	qb *query.Builder
}

type productProfileRepository struct {
	qb *query.Builder
}

type contentSnippetRepository struct {
	qb *query.Builder
}

func NewBrandProfileRepository(qb *query.Builder) domain.BrandProfileRepository {
	return &brandProfileRepository{qb: qb}
}

func NewProductProfileRepository(qb *query.Builder) domain.ProductProfileRepository {
	return &productProfileRepository{qb: qb}
}

func NewContentSnippetRepository(qb *query.Builder) domain.ContentSnippetRepository {
	return &contentSnippetRepository{qb: qb}
}

func (r *brandProfileRepository) Upsert(ctx context.Context, profile *domain.BrandProfile) (*domain.BrandProfile, error) {
	if profile.ProjectID == uuid.Nil {
		return nil, domain.ErrBadRequest.WithMessage("project id is required")
	}
	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}

	now := time.Now()

	query := r.qb.Insert("brand_profiles").
		Columns("id", "project_id", "brand_name", "brand_tone", "font", "style_preset", "brand_colors", "preferred_words", "forbidden_words", "guidelines", "created_at", "updated_at").
		Values(profile.ID, profile.ProjectID, profile.BrandName, profile.BrandTone, profile.Font, profile.StylePreset, toJSON(profile.BrandColors), toJSON(profile.PreferredWords), toJSON(profile.ForbiddenWords), toJSON(profile.Guidelines), now, now).
		Suffix("ON CONFLICT (project_id) DO UPDATE SET brand_name = EXCLUDED.brand_name, brand_tone = EXCLUDED.brand_tone, font = EXCLUDED.font, style_preset = EXCLUDED.style_preset, brand_colors = EXCLUDED.brand_colors, preferred_words = EXCLUDED.preferred_words, forbidden_words = EXCLUDED.forbidden_words, guidelines = EXCLUDED.guidelines, updated_at = NOW() RETURNING id, project_id, brand_name, brand_tone, font, style_preset, brand_colors, preferred_words, forbidden_words, guidelines, created_at, updated_at")

	row := r.qb.QueryRow(query)
	return scanBrandProfile(row)
}

func (r *brandProfileRepository) GetByProjectID(ctx context.Context, projectID string) (*domain.BrandProfile, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid project id")
	}

	query := r.qb.Select("id", "project_id", "brand_name", "brand_tone", "font", "style_preset", "brand_colors", "preferred_words", "forbidden_words", "guidelines", "created_at", "updated_at").
		From("brand_profiles").
		Where(squirrel.Eq{"project_id": pid})

	row := r.qb.QueryRow(query)
	profile, err := scanBrandProfile(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound.WithMessage("brand profile not found")
		}
		return nil, err
	}
	return profile, nil
}

func (r *productProfileRepository) Upsert(ctx context.Context, profile *domain.ProductProfile) (*domain.ProductProfile, error) {
	if profile.ProjectID == uuid.Nil {
		return nil, domain.ErrBadRequest.WithMessage("project id is required")
	}
	if profile.ID == uuid.Nil {
		profile.ID = uuid.New()
	}

	now := time.Now()

	query := r.qb.Insert("product_profiles").
		Columns("id", "project_id", "product_name", "target_audience", "goal", "value_prop", "differentiators", "features", "pricing", "primary_link", "payment_url", "created_at", "updated_at").
		Values(profile.ID, profile.ProjectID, profile.ProductName, profile.TargetAudience, profile.Goal, profile.ValueProp, toJSON(profile.Differentiators), toJSON(profile.Features), toJSON(profile.Pricing), profile.PrimaryLink, profile.PaymentURL, now, now).
		Suffix("ON CONFLICT (project_id) DO UPDATE SET product_name = EXCLUDED.product_name, target_audience = EXCLUDED.target_audience, goal = EXCLUDED.goal, value_prop = EXCLUDED.value_prop, differentiators = EXCLUDED.differentiators, features = EXCLUDED.features, pricing = EXCLUDED.pricing, primary_link = EXCLUDED.primary_link, payment_url = EXCLUDED.payment_url, updated_at = NOW() RETURNING id, project_id, product_name, target_audience, goal, value_prop, differentiators, features, pricing, primary_link, payment_url, created_at, updated_at")

	row := r.qb.QueryRow(query)
	return scanProductProfile(row)
}

func (r *productProfileRepository) GetByProjectID(ctx context.Context, projectID string) (*domain.ProductProfile, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid project id")
	}

	query := r.qb.Select("id", "project_id", "product_name", "target_audience", "goal", "value_prop", "differentiators", "features", "pricing", "primary_link", "payment_url", "created_at", "updated_at").
		From("product_profiles").
		Where(squirrel.Eq{"project_id": pid})

	row := r.qb.QueryRow(query)
	profile, err := scanProductProfile(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound.WithMessage("product profile not found")
		}
		return nil, err
	}
	return profile, nil
}

func (r *contentSnippetRepository) Create(ctx context.Context, snippet *domain.ContentSnippet) (*domain.ContentSnippet, error) {
	if snippet.ProjectID == uuid.Nil {
		return nil, domain.ErrBadRequest.WithMessage("project id is required")
	}
	if snippet.ID == uuid.Nil {
		snippet.ID = uuid.New()
	}
	if snippet.Locale == "" {
		snippet.Locale = "ru"
	}

	query := r.qb.Insert("content_snippets").
		Columns("id", "project_id", "label", "content", "locale", "tags").
		Values(snippet.ID, snippet.ProjectID, snippet.Label, snippet.Content, snippet.Locale, toJSON(snippet.Tags)).
		Suffix("ON CONFLICT (project_id, label, locale) DO UPDATE SET content = EXCLUDED.content, tags = EXCLUDED.tags, created_at = content_snippets.created_at RETURNING id, project_id, label, content, locale, tags, created_at")

	row := r.qb.QueryRow(query)
	return scanContentSnippet(row)
}

func (r *contentSnippetRepository) ListByProject(ctx context.Context, projectID string) ([]*domain.ContentSnippet, error) {
	pid, err := uuid.Parse(projectID)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid project id")
	}

	query := r.qb.Select("id", "project_id", "label", "content", "locale", "tags", "created_at").
		From("content_snippets").
		Where(squirrel.Eq{"project_id": pid}).
		OrderBy("created_at DESC")

	rows, err := r.qb.Query(query)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}
	defer rows.Close()

	var snippets []*domain.ContentSnippet
	for rows.Next() {
		snippet, err := scanContentSnippetRow(rows)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, snippet)
	}

	return snippets, nil
}

func scanBrandProfile(row *sql.Row) (*domain.BrandProfile, error) {
	var profile domain.BrandProfile
	var colorsJSON, preferredJSON, forbiddenJSON, guidelinesJSON []byte
	err := row.Scan(&profile.ID, &profile.ProjectID, &profile.BrandName, &profile.BrandTone, &profile.Font, &profile.StylePreset, &colorsJSON, &preferredJSON, &forbiddenJSON, &guidelinesJSON, &profile.CreatedAt, &profile.UpdatedAt)
	if err != nil {
		return nil, err
	}
	profile.BrandColors = decodeStringArray(colorsJSON)
	profile.PreferredWords = decodeStringArray(preferredJSON)
	profile.ForbiddenWords = decodeStringArray(forbiddenJSON)
	profile.Guidelines = decodeStringMap(guidelinesJSON)
	return &profile, nil
}

func scanProductProfile(row *sql.Row) (*domain.ProductProfile, error) {
	var profile domain.ProductProfile
	var diffJSON, featuresJSON, pricingJSON []byte
	err := row.Scan(&profile.ID, &profile.ProjectID, &profile.ProductName, &profile.TargetAudience, &profile.Goal, &profile.ValueProp, &diffJSON, &featuresJSON, &pricingJSON, &profile.PrimaryLink, &profile.PaymentURL, &profile.CreatedAt, &profile.UpdatedAt)
	if err != nil {
		return nil, err
	}
	profile.Differentiators = decodeStringArray(diffJSON)
	_ = json.Unmarshal(featuresJSON, &profile.Features)
	_ = json.Unmarshal(pricingJSON, &profile.Pricing)
	return &profile, nil
}

func scanContentSnippet(row *sql.Row) (*domain.ContentSnippet, error) {
	var snippet domain.ContentSnippet
	var tagsJSON []byte
	err := row.Scan(&snippet.ID, &snippet.ProjectID, &snippet.Label, &snippet.Content, &snippet.Locale, &tagsJSON, &snippet.CreatedAt)
	if err != nil {
		return nil, err
	}
	snippet.Tags = decodeStringArray(tagsJSON)
	return &snippet, nil
}

func scanContentSnippetRow(row *sql.Rows) (*domain.ContentSnippet, error) {
	var snippet domain.ContentSnippet
	var tagsJSON []byte
	if err := row.Scan(&snippet.ID, &snippet.ProjectID, &snippet.Label, &snippet.Content, &snippet.Locale, &tagsJSON, &snippet.CreatedAt); err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}
	snippet.Tags = decodeStringArray(tagsJSON)
	return &snippet, nil
}

func toJSON(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return string(b)
}

func decodeStringArray(data []byte) []string {
	if len(data) == 0 {
		return nil
	}
	var out []string
	if err := json.Unmarshal(data, &out); err != nil {
		return nil
	}
	return out
}

func decodeStringMap(data []byte) map[string]string {
	if len(data) == 0 {
		return nil
	}
	var out map[string]string
	if err := json.Unmarshal(data, &out); err != nil {
		return nil
	}
	return out
}
