package repositories

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/query"
)

// PublishTargetRepository интерфейс репозитория целей публикации
type PublishTargetRepository interface {
	Create(ctx context.Context, target *domain.PublishTarget) error
	GetByID(ctx context.Context, id string) (*domain.PublishTarget, error)
	GetByProjectID(ctx context.Context, projectID string) (*domain.PublishTarget, error)
	GetBySubdomain(ctx context.Context, subdomain string) (*domain.PublishTarget, error)
	Update(ctx context.Context, target *domain.PublishTarget) error
	Delete(ctx context.Context, id string) error
}

// publishTargetRepository реализация репозитория целей публикации
type publishTargetRepository struct {
	qb *query.Builder
}

// NewPublishTargetRepository создает новый репозиторий целей публикации
func NewPublishTargetRepository(qb *query.Builder) PublishTargetRepository {
	return &publishTargetRepository{qb: qb}
}

// Create создает цель публикации
func (r *publishTargetRepository) Create(ctx context.Context, target *domain.PublishTarget) error {
	query := r.qb.Insert("publish_targets").
		Columns("id", "project_id", "subdomain", "status", "last_published_at", "created_at", "updated_at").
		Values(target.ID, target.ProjectID, target.Subdomain, target.Status, target.LastPublishedAt, target.CreatedAt, target.UpdatedAt)

	_, err := r.qb.Execute(query)
	return err
}

// GetByID получает цель по ID
func (r *publishTargetRepository) GetByID(ctx context.Context, id string) (*domain.PublishTarget, error) {
	targetID, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid target ID format")
	}

	query := r.qb.Select("id", "project_id", "subdomain", "status", "last_published_at", "created_at", "updated_at").
		From("publish_targets").
		Where(squirrel.Eq{"id": targetID})

	row := r.qb.QueryRow(query)

	var target domain.PublishTarget
	err = row.Scan(&target.ID, &target.ProjectID, &target.Subdomain, &target.Status, &target.LastPublishedAt, &target.CreatedAt, &target.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound.WithMessage("target not found")
		}
		return nil, domain.ErrInternal.WithError(err)
	}

	return &target, nil
}

// GetByProjectID получает цель по ID проекта
func (r *publishTargetRepository) GetByProjectID(ctx context.Context, projectID string) (*domain.PublishTarget, error) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid project ID format")
	}

	query := r.qb.Select("id", "project_id", "subdomain", "status", "last_published_at", "created_at", "updated_at").
		From("publish_targets").
		Where(squirrel.Eq{"project_id": projectUUID}).
		OrderBy("updated_at DESC").
		Limit(1)

	row := r.qb.QueryRow(query)

	var target domain.PublishTarget
	err = row.Scan(&target.ID, &target.ProjectID, &target.Subdomain, &target.Status, &target.LastPublishedAt, &target.CreatedAt, &target.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound.WithMessage("target not found")
		}
		return nil, domain.ErrInternal.WithError(err)
	}

	return &target, nil
}

// GetBySubdomain получает цель по поддомену
func (r *publishTargetRepository) GetBySubdomain(ctx context.Context, subdomain string) (*domain.PublishTarget, error) {
	query := r.qb.Select("id", "project_id", "subdomain", "status", "last_published_at", "created_at", "updated_at").
		From("publish_targets").
		Where(squirrel.Eq{"subdomain": subdomain})

	row := r.qb.QueryRow(query)

	var target domain.PublishTarget
	err := row.Scan(&target.ID, &target.ProjectID, &target.Subdomain, &target.Status, &target.LastPublishedAt, &target.CreatedAt, &target.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound.WithMessage("target not found")
		}
		return nil, domain.ErrInternal.WithError(err)
	}

	return &target, nil
}

// Update обновляет цель
func (r *publishTargetRepository) Update(ctx context.Context, target *domain.PublishTarget) error {
	query := r.qb.Update("publish_targets").
		Set("status", target.Status).
		Set("subdomain", target.Subdomain).
		Set("last_published_at", target.LastPublishedAt).
		Set("updated_at", target.UpdatedAt).
		Where(squirrel.Eq{"id": target.ID})

	_, err := r.qb.Execute(query)
	return err
}

// Delete удаляет цель
func (r *publishTargetRepository) Delete(ctx context.Context, id string) error {
	targetID, err := uuid.Parse(id)
	if err != nil {
		return domain.ErrBadRequest.WithMessage("invalid target ID format")
	}

	query := r.qb.Delete("publish_targets").
		Where(squirrel.Eq{"id": targetID})

	_, err = r.qb.Execute(query)
	return err
}
