package repositories

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/query"
)

// schemaVersionRepository реализация репозитория версий схемы
type schemaVersionRepository struct {
	qb *query.Builder
}

// NewSchemaVersionRepository создаёт новый репозиторий версий схемы
func NewSchemaVersionRepository(qb *query.Builder) domain.SchemaVersionRepository {
	return &schemaVersionRepository{qb: qb}
}

// Create создаёт версию схемы
func (r *schemaVersionRepository) Create(ctx context.Context, version *domain.SchemaVersion) error {
	query := r.qb.Insert("project_schema_versions").
		Columns("id", "project_id", "schema_json", "created_at", "created_by", "source", "tokens_used").
		Values(version.ID, version.ProjectID, version.SchemaJSON, version.CreatedAt, version.CreatedBy, version.Source, version.TokensUsed)

	_, err := r.qb.Execute(query)
	return err
}

// GetByID получает версию схемы по ID
func (r *schemaVersionRepository) GetByID(ctx context.Context, id string) (*domain.SchemaVersion, error) {
	versionID, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid version ID format")
	}

	query := r.qb.Select("id", "project_id", "schema_json", "created_at", "created_by", "source", "tokens_used").
		From("project_schema_versions").
		Where(squirrel.Eq{"id": versionID})

	row := r.qb.QueryRow(query)

	var version domain.SchemaVersion
	err = row.Scan(&version.ID, &version.ProjectID, &version.SchemaJSON, &version.CreatedAt, &version.CreatedBy, &version.Source, &version.TokensUsed)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound.WithMessage("schema version not found")
		}
		return nil, domain.ErrInternal.WithError(err)
	}

	return &version, nil
}

// GetByProjectID получает версии схемы проекта
func (r *schemaVersionRepository) GetByProjectID(ctx context.Context, projectID string, limit int) ([]*domain.SchemaVersion, error) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid project ID format")
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := r.qb.Select("id", "project_id", "schema_json", "created_at", "created_by", "source", "tokens_used").
		From("project_schema_versions").
		Where(squirrel.Eq{"project_id": projectUUID}).
		OrderBy("created_at DESC").
		Limit(uint64(limit))

	rows, err := r.qb.Query(query)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}
	defer rows.Close()

	var versions []*domain.SchemaVersion
	for rows.Next() {
		var version domain.SchemaVersion
		err := rows.Scan(&version.ID, &version.ProjectID, &version.SchemaJSON, &version.CreatedAt, &version.CreatedBy, &version.Source, &version.TokensUsed)
		if err != nil {
			return nil, domain.ErrInternal.WithError(err)
		}
		versions = append(versions, &version)
	}

	return versions, nil
}

// Delete удаляет версию схемы
func (r *schemaVersionRepository) Delete(ctx context.Context, id string) error {
	versionID, err := uuid.Parse(id)
	if err != nil {
		return domain.ErrBadRequest.WithMessage("invalid version ID format")
	}

	query := r.qb.Delete("project_schema_versions").
		Where(squirrel.Eq{"id": versionID})

	_, err = r.qb.Execute(query)
	return err
}

