package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/landly/backend/internal/logger"
	domain "github.com/landly/backend/internal/models"
	"go.uber.org/zap"
)

// SchemaVersionService сервис для управления версиями схемы
type SchemaVersionService struct {
	versionRepo domain.SchemaVersionRepository
	projectRepo domain.ProjectRepository
}

// NewSchemaVersionService создаёт новый сервис версий схемы
func NewSchemaVersionService(versionRepo domain.SchemaVersionRepository, projectRepo domain.ProjectRepository) *SchemaVersionService {
	return &SchemaVersionService{
		versionRepo: versionRepo,
		projectRepo: projectRepo,
	}
}

// ListVersions получает список версий схемы проекта
func (s *SchemaVersionService) ListVersions(ctx context.Context, userID, projectID string, limit int) ([]*domain.SchemaVersion, error) {
	// Проверяем доступ к проекту
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, domain.ErrNotFound.WithMessage("project not found")
	}

	if project.UserID.String() != userID {
		return nil, domain.ErrForbidden.WithMessage("access denied")
	}

	if limit <= 0 {
		limit = 20
	}

	versions, err := s.versionRepo.GetByProjectID(ctx, projectID, limit)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	return versions, nil
}

// RevertToVersion восстанавливает схему проекта из версии
func (s *SchemaVersionService) RevertToVersion(ctx context.Context, userID, projectID, versionID string) (*domain.Project, error) {
	// Проверяем доступ к проекту
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, domain.ErrNotFound.WithMessage("project not found")
	}

	if project.UserID.String() != userID {
		return nil, domain.ErrForbidden.WithMessage("access denied")
	}

	// Получаем версию
	version, err := s.versionRepo.GetByID(ctx, versionID)
	if err != nil {
		return nil, domain.ErrNotFound.WithMessage("schema version not found")
	}

	// Проверяем, что версия принадлежит проекту
	if version.ProjectID.String() != projectID {
		return nil, domain.ErrForbidden.WithMessage("version does not belong to this project")
	}

	log := logger.WithContext(ctx).With(
		zap.String("project_id", projectID),
		zap.String("user_id", userID),
		zap.String("version_id", versionID),
	)

	// Сохраняем текущую схему как версию перед откатом
	currentVersion := domain.NewSchemaVersion(
		project.ID,
		uuid.MustParse(userID),
		project.SchemaJSON,
		domain.SchemaVersionSourceSystem,
	)
	if err := s.versionRepo.Create(ctx, currentVersion); err != nil {
		log.Error("failed to save current schema as version before revert", zap.Error(err))
		// Продолжаем, даже если не удалось сохранить версию
	}

	// Восстанавливаем схему из версии
	project.SchemaJSON = version.SchemaJSON
	if err := s.projectRepo.Update(ctx, project); err != nil {
		log.Error("failed to update project schema", zap.Error(err))
		return nil, domain.ErrInternal.WithError(err)
	}

	// Создаём новую версию с source="revert"
	revertVersion := domain.NewSchemaVersion(
		project.ID,
		uuid.MustParse(userID),
		version.SchemaJSON,
		domain.SchemaVersionSourceRevert,
	)
	if err := s.versionRepo.Create(ctx, revertVersion); err != nil {
		log.Error("failed to save revert version", zap.Error(err))
		// Продолжаем, даже если не удалось сохранить версию
	}

	log.Info("schema reverted successfully")

	return project, nil
}

