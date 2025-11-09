package services

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	domain "github.com/landly/backend/internal/models"
)

var defaultPalette = []string{"#1D4ED8", "#111827", "#22C55E"}

type CCEService struct {
	projectRepo domain.ProjectRepository
	brandRepo   domain.BrandProfileRepository
	productRepo domain.ProductProfileRepository
	snippetRepo domain.ContentSnippetRepository
}

func NewCCEService(projectRepo domain.ProjectRepository, brandRepo domain.BrandProfileRepository, productRepo domain.ProductProfileRepository, snippetRepo domain.ContentSnippetRepository) *CCEService {
	return &CCEService{
		projectRepo: projectRepo,
		brandRepo:   brandRepo,
		productRepo: productRepo,
		snippetRepo: snippetRepo,
	}
}

func (s *CCEService) UpsertBrandProfile(ctx context.Context, userID, projectID string, req *domain.UpdateBrandProfileRequest) (*domain.BrandProfile, error) {
	project, err := s.ensureProjectOwnership(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	profile := &domain.BrandProfile{
		ProjectID:      project.ID,
		BrandName:      strings.TrimSpace(req.Name),
		BrandTone:      strings.TrimSpace(req.Tone),
		Font:           strings.TrimSpace(req.Font),
		StylePreset:    strings.TrimSpace(req.StylePreset),
		BrandColors:    sanitizeColors(req.Colors),
		PreferredWords: sanitizeWordList(req.PreferredWords, 16),
		ForbiddenWords: sanitizeWordList(req.ForbiddenWords, 16),
		Guidelines:     sanitizeGuidelines(req.Guidelines),
	}

	updated, err := s.brandRepo.Upsert(ctx, profile)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	return updated, nil
}

func (s *CCEService) UpsertProductProfile(ctx context.Context, userID, projectID string, req *domain.UpdateProductProfileRequest) (*domain.ProductProfile, error) {
	project, err := s.ensureProjectOwnership(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	profile := &domain.ProductProfile{
		ProjectID:       project.ID,
		ProductName:     strings.TrimSpace(req.Name),
		TargetAudience:  strings.TrimSpace(req.TargetAudience),
		Goal:            strings.TrimSpace(req.Goal),
		ValueProp:       strings.TrimSpace(req.ValueProp),
		Differentiators: sanitizeWordList(req.Differentiators, 20),
		Features:        sanitizeFeatures(req.Features),
		Pricing:         sanitizePlans(req.Pricing),
		PrimaryLink:     strings.TrimSpace(req.PrimaryLink),
		PaymentURL:      strings.TrimSpace(req.PaymentURL),
	}

	updated, err := s.productRepo.Upsert(ctx, profile)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	return updated, nil
}

func (s *CCEService) CreateSnippet(ctx context.Context, userID, projectID string, req *domain.CreateSnippetRequest) (*domain.ContentSnippet, error) {
	project, err := s.ensureProjectOwnership(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	snippet := &domain.ContentSnippet{
		ProjectID: project.ID,
		Label:     strings.TrimSpace(req.Label),
		Content:   strings.TrimSpace(req.Content),
		Locale:    strings.TrimSpace(req.Locale),
		Tags:      sanitizeWordList(req.Tags, 12),
	}

	created, err := s.snippetRepo.Create(ctx, snippet)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	return created, nil
}

func (s *CCEService) ListSnippets(ctx context.Context, userID, projectID string) ([]*domain.ContentSnippet, error) {
	if _, err := s.ensureProjectOwnership(ctx, userID, projectID); err != nil {
		return nil, err
	}

	snippets, err := s.snippetRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	return snippets, nil
}

func (s *CCEService) BuildPromptContext(ctx context.Context, userID, projectID string) (*PromptContext, error) {
	project, err := s.ensureProjectOwnership(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}

	var brand *domain.BrandProfile
	if bp, err := s.brandRepo.GetByProjectID(ctx, projectID); err == nil {
		brand = bp
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, domain.ErrInternal.WithError(err)
	}

	var product *domain.ProductProfile
	if pp, err := s.productRepo.GetByProjectID(ctx, projectID); err == nil {
		product = pp
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, domain.ErrInternal.WithError(err)
	}

	snippets, err := s.snippetRepo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	context := &PromptContext{
		Project: ProjectContext{
			ID:            project.ID,
			Name:          project.Name,
			Niche:         project.Niche,
			SchemaVersion: project.SchemaVersion,
		},
		Brand:      buildBrandContext(project, brand),
		Product:    buildProductContext(project, product),
		Content:    ContentContext{Snippets: snippets},
		Generation: buildGenerationContext(brand, product),
	}

	return context, nil
}

func (s *CCEService) ensureProjectOwnership(ctx context.Context, userID, projectID string) (*domain.Project, error) {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, domain.ErrNotFound.WithMessage("project not found")
	}

	if project.UserID.String() != userID {
		return nil, domain.ErrForbidden.WithMessage("access denied")
	}

	return project, nil
}

func buildBrandContext(project *domain.Project, profile *domain.BrandProfile) BrandContext {
	ctx := BrandContext{
		Name:   project.Name,
		Tone:   "neutral",
		Colors: append([]string{}, defaultPalette...),
		Font:   "Inter",
	}

	if profile == nil {
		return ctx
	}

	if profile.BrandName != "" {
		ctx.Name = profile.BrandName
	}
	if profile.BrandTone != "" {
		ctx.Tone = profile.BrandTone
	}
	if len(profile.BrandColors) > 0 {
		ctx.Colors = profile.BrandColors
	}
	if profile.Font != "" {
		ctx.Font = profile.Font
	}
	ctx.StylePreset = profile.StylePreset
	ctx.PreferredWords = profile.PreferredWords
	ctx.ForbiddenWords = profile.ForbiddenWords
	ctx.Guidelines = profile.Guidelines

	return ctx
}

func buildProductContext(project *domain.Project, profile *domain.ProductProfile) ProductContext {
	ctx := ProductContext{
		Name:           project.Name,
		TargetAudience: project.Niche,
	}

	if profile == nil {
		return ctx
	}

	if profile.ProductName != "" {
		ctx.Name = profile.ProductName
	}
	if profile.TargetAudience != "" {
		ctx.TargetAudience = profile.TargetAudience
	}
	ctx.Goal = profile.Goal
	ctx.ValueProp = profile.ValueProp
	ctx.Differentiators = profile.Differentiators
	ctx.Features = profile.Features
	ctx.Pricing = profile.Pricing
	ctx.PrimaryLink = profile.PrimaryLink
	ctx.PaymentURL = profile.PaymentURL

	return ctx
}

func buildGenerationContext(brand *domain.BrandProfile, product *domain.ProductProfile) GenerationContext {
	ctx := GenerationContext{
		Goal:     "leadgen",
		Tone:     "neutral",
		Language: "ru",
	}

	if brand != nil && brand.BrandTone != "" {
		ctx.Tone = brand.BrandTone
	}
	if product != nil {
		if product.Goal != "" {
			ctx.Goal = product.Goal
		}
		if product.PaymentURL != "" {
			ctx.PaymentURL = product.PaymentURL
		}
	}

	return ctx
}

func sanitizeColors(colors []string) []string {
	if len(colors) == 0 {
		return nil
	}
	result := make([]string, 0, len(colors))
	seen := map[string]struct{}{}
	for _, color := range colors {
		trimmed := strings.ToUpper(strings.TrimSpace(color))
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "#") {
			trimmed = "#" + trimmed
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func sanitizeWordList(list []string, limit int) []string {
	if len(list) == 0 {
		return nil
	}
	result := make([]string, 0, len(list))
	seen := map[string]struct{}{}
	for _, word := range list {
		trimmed := strings.TrimSpace(word)
		if trimmed == "" {
			continue
		}
		lower := strings.ToLower(trimmed)
		if _, ok := seen[lower]; ok {
			continue
		}
		seen[lower] = struct{}{}
		result = append(result, trimmed)
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func sanitizeGuidelines(src map[string]string) map[string]string {
	if len(src) == 0 {
		return nil
	}
	result := make(map[string]string, len(src))
	for key, value := range src {
		trimmedKey := strings.TrimSpace(key)
		trimmedValue := strings.TrimSpace(value)
		if trimmedKey == "" || trimmedValue == "" {
			continue
		}
		result[trimmedKey] = trimmedValue
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func sanitizeFeatures(features []domain.ProductFeature) []domain.ProductFeature {
	if len(features) == 0 {
		return nil
	}
	result := make([]domain.ProductFeature, 0, len(features))
	for _, feature := range features {
		title := strings.TrimSpace(feature.Title)
		description := strings.TrimSpace(feature.Description)
		if title == "" && description == "" {
			continue
		}
		result = append(result, domain.ProductFeature{Title: title, Description: description})
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func sanitizePlans(plans []domain.ProductPlan) []domain.ProductPlan {
	if len(plans) == 0 {
		return nil
	}
	result := make([]domain.ProductPlan, 0, len(plans))
	for _, plan := range plans {
		name := strings.TrimSpace(plan.Name)
		if name == "" {
			continue
		}
		result = append(result, domain.ProductPlan{
			Name:       name,
			Price:      strings.TrimSpace(plan.Price),
			Currency:   strings.TrimSpace(plan.Currency),
			Period:     strings.TrimSpace(plan.Period),
			Features:   sanitizeWordList(plan.Features, 10),
			ButtonText: strings.TrimSpace(plan.ButtonText),
			URL:        strings.TrimSpace(plan.URL),
			Featured:   plan.Featured,
		})
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// Context DTOs

type PromptContext struct {
	Project    ProjectContext    `json:"project"`
	Brand      BrandContext      `json:"brand"`
	Product    ProductContext    `json:"product"`
	Content    ContentContext    `json:"content"`
	Generation GenerationContext `json:"generation"`
}

type ProjectContext struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Niche         string    `json:"niche"`
	SchemaVersion int       `json:"schema_version"`
}

type BrandContext struct {
	Name           string            `json:"name"`
	Tone           string            `json:"tone"`
	Font           string            `json:"font"`
	StylePreset    string            `json:"style_preset"`
	Colors         []string          `json:"colors"`
	PreferredWords []string          `json:"preferred_words"`
	ForbiddenWords []string          `json:"forbidden_words"`
	Guidelines     map[string]string `json:"guidelines"`
}

type ProductContext struct {
	Name            string                  `json:"name"`
	TargetAudience  string                  `json:"target_audience"`
	Goal            string                  `json:"goal"`
	ValueProp       string                  `json:"value_prop"`
	Differentiators []string                `json:"differentiators"`
	Features        []domain.ProductFeature `json:"features"`
	Pricing         []domain.ProductPlan    `json:"pricing"`
	PrimaryLink     string                  `json:"primary_link"`
	PaymentURL      string                  `json:"payment_url"`
}

type ContentContext struct {
	Snippets []*domain.ContentSnippet `json:"snippets"`
}

type GenerationContext struct {
	Goal       string `json:"goal"`
	Tone       string `json:"tone"`
	Language   string `json:"language"`
	PaymentURL string `json:"payment_url,omitempty"`
}
