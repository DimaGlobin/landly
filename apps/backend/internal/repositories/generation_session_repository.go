package repositories

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/query"
)

// GenerationSessionRepository интерфейс репозитория сессий генерации
type GenerationSessionRepository interface {
	Create(ctx context.Context, session *domain.GenerationSession) error
	GetByID(ctx context.Context, id string) (*domain.GenerationSession, error)
	GetByProjectID(ctx context.Context, projectID string) ([]*domain.GenerationSession, error)
	Update(ctx context.Context, session *domain.GenerationSession) error
	Delete(ctx context.Context, id string) error
}

// generationSessionRepository реализация репозитория сессий генерации
type generationSessionRepository struct {
	qb *query.Builder
}

// NewGenerationSessionRepository создает новый репозиторий сессий генерации
func NewGenerationSessionRepository(qb *query.Builder) GenerationSessionRepository {
	return &generationSessionRepository{qb: qb}
}

// Create создает сессию генерации
func (r *generationSessionRepository) Create(ctx context.Context, session *domain.GenerationSession) error {
	query := r.qb.Insert("generation_sessions").
		Columns("id", "project_id", "prompt", "model", "status", "schema_json", "completed_at", "created_at", "updated_at").
		Values(session.ID, session.ProjectID, session.Prompt, session.Model, session.Status, session.SchemaJSON, session.CompletedAt, session.CreatedAt, session.UpdatedAt)

	_, err := r.qb.Execute(query)
	return err
}

// GetByID получает сессию по ID
func (r *generationSessionRepository) GetByID(ctx context.Context, id string) (*domain.GenerationSession, error) {
	sessionID, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid session ID format")
	}

	query := r.qb.Select("id", "project_id", "prompt", "model", "status", "schema_json", "completed_at", "created_at", "updated_at").
		From("generation_sessions").
		Where(squirrel.Eq{"id": sessionID})

	row := r.qb.QueryRow(query)

	var session domain.GenerationSession
	err = row.Scan(&session.ID, &session.ProjectID, &session.Prompt, &session.Model, &session.Status, &session.SchemaJSON, &session.CompletedAt, &session.CreatedAt, &session.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound.WithMessage("session not found")
		}
		return nil, domain.ErrInternal.WithError(err)
	}

	return &session, nil
}

// GetByProjectID получает сессии проекта
func (r *generationSessionRepository) GetByProjectID(ctx context.Context, projectID string) ([]*domain.GenerationSession, error) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid project ID format")
	}

	query := r.qb.Select("id", "project_id", "prompt", "model", "status", "schema_json", "completed_at", "created_at", "updated_at").
		From("generation_sessions").
		Where(squirrel.Eq{"project_id": projectUUID}).
		OrderBy("created_at DESC")

	rows, err := r.qb.Query(query)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}
	defer rows.Close()

	var sessions []*domain.GenerationSession
	for rows.Next() {
		var session domain.GenerationSession
		err := rows.Scan(&session.ID, &session.ProjectID, &session.Prompt, &session.Model, &session.Status, &session.SchemaJSON, &session.CompletedAt, &session.CreatedAt, &session.UpdatedAt)
		if err != nil {
			return nil, domain.ErrInternal.WithError(err)
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// Update обновляет сессию
func (r *generationSessionRepository) Update(ctx context.Context, session *domain.GenerationSession) error {
	query := r.qb.Update("generation_sessions").
		Set("prompt", session.Prompt).
		Set("status", session.Status).
		Set("schema_json", session.SchemaJSON).
		Set("completed_at", session.CompletedAt).
		Set("updated_at", session.UpdatedAt).
		Where(squirrel.Eq{"id": session.ID})

	_, err := r.qb.Execute(query)
	return err
}

// Delete удаляет сессию
func (r *generationSessionRepository) Delete(ctx context.Context, id string) error {
	sessionID, err := uuid.Parse(id)
	if err != nil {
		return domain.ErrBadRequest.WithMessage("invalid session ID format")
	}

	query := r.qb.Delete("generation_sessions").
		Where(squirrel.Eq{"id": sessionID})

	_, err = r.qb.Execute(query)
	return err
}
