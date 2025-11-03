package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/query"
)

// ProjectRepository интерфейс репозитория проектов
type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) error
	GetByID(ctx context.Context, id string) (*domain.Project, error)
	GetByUserID(ctx context.Context, userID string) ([]*domain.Project, error)
	Update(ctx context.Context, project *domain.Project) error
	Delete(ctx context.Context, id string) error
	UpdateSchema(ctx context.Context, projectID string, schemaJSON string) error
}

// projectRepository реализация репозитория проектов
type projectRepository struct {
	qb *query.Builder
}

// NewProjectRepository создает новый репозиторий проектов
func NewProjectRepository(qb *query.Builder) ProjectRepository {
	return &projectRepository{qb: qb}
}

// Create создает проект
func (r *projectRepository) Create(ctx context.Context, project *domain.Project) error {
	query := r.qb.Insert("projects").
		Columns("id", "user_id", "name", "niche", "schema_json", "status", "created_at", "updated_at").
		Values(project.ID, project.UserID, project.Name, project.Niche, project.SchemaJSON, project.Status, project.CreatedAt, project.UpdatedAt)

	_, err := r.qb.Execute(query)
	return err
}

// GetByID получает проект по ID
func (r *projectRepository) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	projectID, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid project ID format")
	}

	query := r.qb.Select("id", "user_id", "name", "niche", "schema_json", "status", "created_at", "updated_at").
		From("projects").
		Where(squirrel.Eq{"id": projectID})

	row := r.qb.QueryRow(query)

	var project domain.Project
	err = row.Scan(&project.ID, &project.UserID, &project.Name, &project.Niche, &project.SchemaJSON, &project.Status, &project.CreatedAt, &project.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound.WithMessage("project not found")
		}
		return nil, domain.ErrInternal.WithError(err)
	}

	return &project, nil
}

// GetByUserID получает проекты пользователя
func (r *projectRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Project, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid user ID format")
	}

	query := r.qb.Select("id", "user_id", "name", "niche", "schema_json", "status", "created_at", "updated_at").
		From("projects").
		Where(squirrel.Eq{"user_id": userUUID}).
		OrderBy("updated_at DESC")

	rows, err := r.qb.Query(query)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		var project domain.Project
		err := rows.Scan(&project.ID, &project.UserID, &project.Name, &project.Niche, &project.SchemaJSON, &project.Status, &project.CreatedAt, &project.UpdatedAt)
		if err != nil {
			return nil, domain.ErrInternal.WithError(err)
		}
		projects = append(projects, &project)
	}

	return projects, nil
}

// Update обновляет проект
func (r *projectRepository) Update(ctx context.Context, project *domain.Project) error {
	now := time.Now()
	project.UpdatedAt = now

	query := r.qb.Update("projects").
		Set("name", project.Name).
		Set("niche", project.Niche).
		Set("schema_json", project.SchemaJSON).
		Set("status", project.Status).
		Set("updated_at", project.UpdatedAt).
		Where(squirrel.Eq{"id": project.ID})

	_, err := r.qb.Execute(query)
	return err
}

// Delete удаляет проект
func (r *projectRepository) Delete(ctx context.Context, id string) error {
	projectID, err := uuid.Parse(id)
	if err != nil {
		return domain.ErrBadRequest.WithMessage("invalid project ID format")
	}

	query := r.qb.Delete("projects").
		Where(squirrel.Eq{"id": projectID})

	_, err = r.qb.Execute(query)
	return err
}

// UpdateSchema обновляет схему проекта
func (r *projectRepository) UpdateSchema(ctx context.Context, projectID string, schemaJSON string) error {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return domain.ErrBadRequest.WithMessage("invalid project ID format")
	}

	now := time.Now()

	query := r.qb.Update("projects").
		Set("schema_json", schemaJSON).
		Set("status", domain.ProjectStatusGenerated).
		Set("updated_at", now).
		Where(squirrel.Eq{"id": projectUUID})

	_, err = r.qb.Execute(query)
	return err
}
