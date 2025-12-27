package repositories

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/query"
)

// PublishJobRepository интерфейс репозитория задач публикации
type PublishJobRepository interface {
	Create(ctx context.Context, job *domain.PublishJob) error
	GetByID(ctx context.Context, id string) (*domain.PublishJob, error)
	GetByProjectID(ctx context.Context, projectID string) ([]*domain.PublishJob, error)
	GetByStatus(ctx context.Context, status string) ([]*domain.PublishJob, error)
	GetPendingByProjectID(ctx context.Context, projectID string) (*domain.PublishJob, error)
	Update(ctx context.Context, job *domain.PublishJob) error
	Delete(ctx context.Context, id string) error
}

// publishJobRepository реализация репозитория задач публикации
type publishJobRepository struct {
	qb *query.Builder
}

// NewPublishJobRepository создает новый репозиторий задач публикации
func NewPublishJobRepository(qb *query.Builder) PublishJobRepository {
	return &publishJobRepository{qb: qb}
}

// Create создает задачу публикации
func (r *publishJobRepository) Create(ctx context.Context, job *domain.PublishJob) error {
	query := r.qb.Insert("publish_jobs").
		Columns("id", "project_id", "publish_target_id", "status", "release_id", "error_message", "started_at", "completed_at", "created_at", "updated_at").
		Values(job.ID, job.ProjectID, job.PublishTargetID, job.Status, job.ReleaseID, job.ErrorMessage, job.StartedAt, job.CompletedAt, job.CreatedAt, job.UpdatedAt)

	_, err := r.qb.Execute(query)
	return err
}

// GetByID получает задачу по ID
func (r *publishJobRepository) GetByID(ctx context.Context, id string) (*domain.PublishJob, error) {
	jobID, err := uuid.Parse(id)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid job ID format")
	}

	query := r.qb.Select("id", "project_id", "publish_target_id", "status", "release_id", "error_message", "started_at", "completed_at", "created_at", "updated_at").
		From("publish_jobs").
		Where(squirrel.Eq{"id": jobID})

	row := r.qb.QueryRow(query)

	var job domain.PublishJob
	err = row.Scan(&job.ID, &job.ProjectID, &job.PublishTargetID, &job.Status, &job.ReleaseID, &job.ErrorMessage, &job.StartedAt, &job.CompletedAt, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound.WithMessage("job not found")
		}
		return nil, domain.ErrInternal.WithError(err)
	}

	return &job, nil
}

// GetByProjectID получает все задачи по ID проекта
func (r *publishJobRepository) GetByProjectID(ctx context.Context, projectID string) ([]*domain.PublishJob, error) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid project ID format")
	}

	query := r.qb.Select("id", "project_id", "publish_target_id", "status", "release_id", "error_message", "started_at", "completed_at", "created_at", "updated_at").
		From("publish_jobs").
		Where(squirrel.Eq{"project_id": projectUUID}).
		OrderBy("created_at DESC")

	rows, err := r.qb.Query(query)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}
	defer rows.Close()

	var jobs []*domain.PublishJob
	for rows.Next() {
		var job domain.PublishJob
		err := rows.Scan(&job.ID, &job.ProjectID, &job.PublishTargetID, &job.Status, &job.ReleaseID, &job.ErrorMessage, &job.StartedAt, &job.CompletedAt, &job.CreatedAt, &job.UpdatedAt)
		if err != nil {
			return nil, domain.ErrInternal.WithError(err)
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// GetByStatus получает все задачи по статусу
func (r *publishJobRepository) GetByStatus(ctx context.Context, status string) ([]*domain.PublishJob, error) {
	query := r.qb.Select("id", "project_id", "publish_target_id", "status", "release_id", "error_message", "started_at", "completed_at", "created_at", "updated_at").
		From("publish_jobs").
		Where(squirrel.Eq{"status": status}).
		OrderBy("created_at ASC")

	rows, err := r.qb.Query(query)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}
	defer rows.Close()

	var jobs []*domain.PublishJob
	for rows.Next() {
		var job domain.PublishJob
		err := rows.Scan(&job.ID, &job.ProjectID, &job.PublishTargetID, &job.Status, &job.ReleaseID, &job.ErrorMessage, &job.StartedAt, &job.CompletedAt, &job.CreatedAt, &job.UpdatedAt)
		if err != nil {
			return nil, domain.ErrInternal.WithError(err)
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// GetPendingByProjectID получает ожидающую задачу по ID проекта
func (r *publishJobRepository) GetPendingByProjectID(ctx context.Context, projectID string) (*domain.PublishJob, error) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid project ID format")
	}

	query := r.qb.Select("id", "project_id", "publish_target_id", "status", "release_id", "error_message", "started_at", "completed_at", "created_at", "updated_at").
		From("publish_jobs").
		Where(squirrel.Eq{"project_id": projectUUID, "status": domain.PublishJobStatusPending}).
		OrderBy("created_at ASC").
		Limit(1)

	row := r.qb.QueryRow(query)

	var job domain.PublishJob
	err = row.Scan(&job.ID, &job.ProjectID, &job.PublishTargetID, &job.Status, &job.ReleaseID, &job.ErrorMessage, &job.StartedAt, &job.CompletedAt, &job.CreatedAt, &job.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound.WithMessage("no pending job found")
		}
		return nil, domain.ErrInternal.WithError(err)
	}

	return &job, nil
}

// Update обновляет задачу
func (r *publishJobRepository) Update(ctx context.Context, job *domain.PublishJob) error {
	query := r.qb.Update("publish_jobs").
		Set("status", job.Status).
		Set("release_id", job.ReleaseID).
		Set("error_message", job.ErrorMessage).
		Set("started_at", job.StartedAt).
		Set("completed_at", job.CompletedAt).
		Set("updated_at", job.UpdatedAt).
		Where(squirrel.Eq{"id": job.ID})

	_, err := r.qb.Execute(query)
	return err
}

// Delete удаляет задачу
func (r *publishJobRepository) Delete(ctx context.Context, id string) error {
	jobID, err := uuid.Parse(id)
	if err != nil {
		return domain.ErrBadRequest.WithMessage("invalid job ID format")
	}

	query := r.qb.Delete("publish_jobs").
		Where(squirrel.Eq{"id": jobID})

	_, err = r.qb.Execute(query)
	return err
}

