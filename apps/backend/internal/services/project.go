package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/landly/backend/internal/logger"
	"github.com/landly/backend/internal/metrics"
	domain "github.com/landly/backend/internal/models"
	"go.uber.org/zap"
)

// ProjectRepository интерфейс для репозитория проектов
type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) error
	GetByID(ctx context.Context, id string) (*domain.Project, error)
	GetByUserID(ctx context.Context, userID string) ([]*domain.Project, error)
	Update(ctx context.Context, project *domain.Project) error
	Delete(ctx context.Context, id string) error
}

// PublishTargetRepository интерфейс для репозитория целей публикации
type PublishTargetRepository interface {
	GetByProjectID(ctx context.Context, projectID string) (*domain.PublishTarget, error)
}

// ProjectService сервис для управления проектами
type ProjectService struct {
	projectRepo       ProjectRepository
	publishTargetRepo PublishTargetRepository
	publisher         Publisher
}

// NewProjectService создаёт новый project service
func NewProjectService(projectRepo ProjectRepository, publishTargetRepo PublishTargetRepository, publisher Publisher) *ProjectService {
	return &ProjectService{
		projectRepo:       projectRepo,
		publishTargetRepo: publishTargetRepo,
		publisher:         publisher,
	}
}

// CreateProject создаёт новый проект
func (s *ProjectService) CreateProject(ctx context.Context, userID string, req *domain.CreateProjectRequest) (*domain.Project, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid user ID")
	}

	project := domain.NewProject(userUUID, req.Name, req.Niche)

	// Валидация проекта
	if project.Name == "" {
		return nil, domain.ErrBadRequest.WithMessage("project name is required")
	}
	if project.Niche == "" {
		return nil, domain.ErrBadRequest.WithMessage("project niche is required")
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	// Record project creation metric
	metrics.ProjectsCreatedTotal.Inc()

	return project, nil
}

// GetProject получает проект по ID
func (s *ProjectService) GetProject(ctx context.Context, userID, projectID string) (*domain.Project, error) {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, domain.ErrNotFound.WithMessage("project not found")
	}

	// Проверка принадлежности проекта пользователю
	if project.UserID.String() != userID {
		return nil, domain.ErrForbidden.WithMessage("access denied")
	}

	return project, nil
}

// ListProjects получает проекты пользователя
func (s *ProjectService) ListProjects(ctx context.Context, userID string) ([]*domain.Project, error) {
	projects, err := s.projectRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	return projects, nil
}

// UpdateProject обновляет проект
func (s *ProjectService) UpdateProject(ctx context.Context, userID, projectID string, req *domain.UpdateProjectRequest) (*domain.Project, error) {
	// Проверяем существование проекта
	existingProject, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, domain.ErrNotFound.WithMessage("project not found")
	}

	// Проверка принадлежности проекта пользователю
	if existingProject.UserID.String() != userID {
		return nil, domain.ErrForbidden.WithMessage("access denied")
	}

	// Обновляем поля проекта
	if req.Name != "" {
		existingProject.Name = req.Name
	}
	if req.Niche != "" {
		existingProject.Niche = req.Niche
	}

	// Обновляем проект
	if err := s.projectRepo.Update(ctx, existingProject); err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	return existingProject, nil
}

// DeleteProject удаляет проект
func (s *ProjectService) DeleteProject(ctx context.Context, userID, projectID string) error {
	log := logger.WithContext(ctx).With(
		zap.String("project_id", projectID),
		zap.String("user_id", userID),
	)

	// Проверяем существование проекта
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return domain.ErrNotFound.WithMessage("project not found")
	}

	// Проверка принадлежности проекта пользователю
	if project.UserID.String() != userID {
		return domain.ErrForbidden.WithMessage("access denied")
	}

	// Удаляем файлы из S3 перед удалением проекта
	target, err := s.publishTargetRepo.GetByProjectID(ctx, projectID)
	if err == nil && target != nil {
		// Формируем префикс для удаления: sites/{subdomain}/
		prefix := fmt.Sprintf("sites/%s", target.Subdomain)
		log.Info("deleting S3 files for project", zap.String("prefix", prefix))

		if err := s.publisher.DeletePrefix(ctx, prefix); err != nil {
			// Логируем ошибку, но не прерываем удаление проекта
			// Файлы могут быть уже удалены или недоступны
			log.Warn("failed to delete S3 files", zap.String("prefix", prefix), zap.Error(err))
		} else {
			log.Info("successfully deleted S3 files", zap.String("prefix", prefix))
		}
	}

	// Удаляем проект из БД
	if err := s.projectRepo.Delete(ctx, projectID); err != nil {
		return domain.ErrInternal.WithError(err)
	}

	// Record project deletion metric
	metrics.ProjectsDeletedTotal.Inc()

	log.Info("project deleted successfully")
	return nil
}
