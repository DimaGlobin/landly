package repositories

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/query"
)

// integrationRepository реализация репозитория интеграций
type integrationRepository struct {
	qb *query.Builder
}

// NewIntegrationRepository создает новый репозиторий интеграций
func NewIntegrationRepository(qb *query.Builder) domain.IntegrationRepository {
	return &integrationRepository{qb: qb}
}

// Create создает интеграцию
func (r *integrationRepository) Create(ctx context.Context, integration *domain.Integration) error {
	query := r.qb.Insert("integrations").
		Columns("id", "project_id", "type", "config", "created_at", "updated_at").
		Values(integration.ID, integration.ProjectID, integration.Type, integration.Config, integration.CreatedAt, integration.UpdatedAt)

	_, err := r.qb.Execute(query)
	return err
}

// GetByID получает интеграцию по ID
func (r *integrationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Integration, error) {
	query := r.qb.Select("id", "project_id", "type", "config", "created_at", "updated_at").
		From("integrations").
		Where(squirrel.Eq{"id": id})

	row := r.qb.QueryRow(query)

	var integration domain.Integration
	err := row.Scan(&integration.ID, &integration.ProjectID, &integration.Type, &integration.Config, &integration.CreatedAt, &integration.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound.WithMessage("integration not found")
		}
		return nil, domain.ErrInternal.WithError(err)
	}

	return &integration, nil
}

// GetByProjectID получает интеграции проекта
func (r *integrationRepository) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*domain.Integration, error) {
	query := r.qb.Select("id", "project_id", "type", "config", "created_at", "updated_at").
		From("integrations").
		Where(squirrel.Eq{"project_id": projectID}).
		OrderBy("created_at DESC")

	rows, err := r.qb.Query(query)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}
	defer rows.Close()

	var integrations []*domain.Integration
	for rows.Next() {
		var integration domain.Integration
		err := rows.Scan(&integration.ID, &integration.ProjectID, &integration.Type, &integration.Config, &integration.CreatedAt, &integration.UpdatedAt)
		if err != nil {
			return nil, domain.ErrInternal.WithError(err)
		}
		integrations = append(integrations, &integration)
	}

	return integrations, nil
}

// GetByProjectIDAndType получает интеграцию по ID проекта и типу
func (r *integrationRepository) GetByProjectIDAndType(ctx context.Context, projectID uuid.UUID, integrationType string) (*domain.Integration, error) {
	query := r.qb.Select("id", "project_id", "type", "config", "created_at", "updated_at").
		From("integrations").
		Where(squirrel.Eq{"project_id": projectID, "type": integrationType})

	row := r.qb.QueryRow(query)

	var integration domain.Integration
	err := row.Scan(&integration.ID, &integration.ProjectID, &integration.Type, &integration.Config, &integration.CreatedAt, &integration.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound.WithMessage("integration not found")
		}
		return nil, domain.ErrInternal.WithError(err)
	}

	return &integration, nil
}

// Update обновляет интеграцию
func (r *integrationRepository) Update(ctx context.Context, integration *domain.Integration) error {
	query := r.qb.Update("integrations").
		Set("type", integration.Type).
		Set("config", integration.Config).
		Set("updated_at", integration.UpdatedAt).
		Where(squirrel.Eq{"id": integration.ID})

	_, err := r.qb.Execute(query)
	return err
}

// Delete удаляет интеграцию
func (r *integrationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := r.qb.Delete("integrations").
		Where(squirrel.Eq{"id": id})

	_, err := r.qb.Execute(query)
	return err
}
