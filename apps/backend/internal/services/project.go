package services

import (
	"context"

	"github.com/google/uuid"
	domain "github.com/landly/backend/internal/models"
)

// ProjectRepository интерфейс для репозитория проектов
type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) error
	GetByID(ctx context.Context, id string) (*domain.Project, error)
	GetByUserID(ctx context.Context, userID string) ([]*domain.Project, error)
	Update(ctx context.Context, project *domain.Project) error
	Delete(ctx context.Context, id string) error
}

// ProjectService сервис для управления проектами
type ProjectService struct {
	projectRepo ProjectRepository
}

// NewProjectService создаёт новый project service
func NewProjectService(projectRepo ProjectRepository) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
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
	// Проверяем существование проекта
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return domain.ErrNotFound.WithMessage("project not found")
	}

	// Проверка принадлежности проекта пользователю
	if project.UserID.String() != userID {
		return domain.ErrForbidden.WithMessage("access denied")
	}

	// Удаляем проект
	if err := s.projectRepo.Delete(ctx, projectID); err != nil {
		return domain.ErrInternal.WithError(err)
	}

	return nil
}
