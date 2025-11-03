package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/landly/backend/internal/logger"
	domain "github.com/landly/backend/internal/models"
	"go.uber.org/zap"
)

// Renderer интерфейс для рендеринга статических сайтов
type Renderer interface {
	RenderStatic(ctx context.Context, projectID uuid.UUID, schemaJSON string) (string, error)
}

// Publisher интерфейс для публикации в S3/CDN
type Publisher interface {
	Upload(ctx context.Context, localPath, remotePath string) error
	GetPublicURL(remotePath string) string
	GetObject(ctx context.Context, remotePath string) (io.ReadCloser, string, error)
}

// PublishUserRepository интерфейс для получения пользователя
type PublishUserRepository interface {
	GetByID(ctx context.Context, userID string) (*domain.User, error)
}

// PublishService сервис для публикации проектов
type PublishService struct {
	projectRepo       domain.ProjectRepository
	publishTargetRepo domain.PublishTargetRepository
	userRepo          PublishUserRepository
	renderer          Renderer
	publisher         Publisher
	publicBase        string
}

// PublishResult результат публикации
type PublishResult struct {
	Subdomain   string `json:"subdomain"`
	PublicURL   string `json:"public_url"`
	PublishedAt string `json:"published_at"`
}

// NewPublishService создаёт новый publish service
func NewPublishService(
	projectRepo domain.ProjectRepository,
	publishTargetRepo domain.PublishTargetRepository,
	userRepo PublishUserRepository,
	renderer Renderer,
	publisher Publisher,
	publicBase string,
) *PublishService {
	return &PublishService{
		projectRepo:       projectRepo,
		publishTargetRepo: publishTargetRepo,
		userRepo:          userRepo,
		renderer:          renderer,
		publisher:         publisher,
		publicBase:        strings.TrimRight(publicBase, "/"),
	}
}

func generateSubdomain(projectName string, projectID uuid.UUID) string {
	prefix := strings.ToLower(strings.TrimSpace(projectName))
	if prefix == "" {
		prefix = "project"
	}

	var builder strings.Builder
	lastHyphen := false

	for _, r := range prefix {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastHyphen = false
		} else {
			if !lastHyphen {
				builder.WriteRune('-')
				lastHyphen = true
			}
		}
	}

	slug := strings.Trim(builder.String(), "-")
	if slug == "" {
		slug = "project"
	}

	return fmt.Sprintf("%s-%s", slug, projectID.String()[:8])
}

func (s *PublishService) publicBaseURL() string {
	if s.publicBase != "" {
		return s.publicBase
	}
	return "http://localhost:8080"
}

// PublishSite публикует сайт (новый интерфейс)
func (s *PublishService) PublishSite(ctx context.Context, userID, projectID string, req *domain.PublishRequest) (*domain.PublishTarget, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid user ID")
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid project ID")
	}

	// Проверка доступа к проекту
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, domain.ErrNotFound.WithMessage("project not found")
	}

	if project.UserID.String() != userID {
		return nil, domain.ErrForbidden.WithMessage("access denied")
	}

	// Создаем или обновляем цель публикации с использованием имени проекта
	// Формат: <base-url>/sites/<subdomain>
	subdomain := generateSubdomain(project.Name, projectUUID)
	var target *domain.PublishTarget

	existingTarget, err := s.publishTargetRepo.GetByProjectID(ctx, projectUUID.String())
	if err == nil && existingTarget != nil {
		existingTarget.Subdomain = subdomain
		existingTarget.Status = domain.PublishStatusDraft
		existingTarget.UpdatedAt = time.Now()
		if err := s.publishTargetRepo.Update(ctx, existingTarget); err != nil {
			return nil, domain.ErrInternal.WithError(err)
		}
		target = existingTarget
	} else {
		target = domain.NewPublishTarget(projectUUID, subdomain)
		if err := s.publishTargetRepo.Create(ctx, target); err != nil {
			return nil, domain.ErrInternal.WithError(err)
		}
	}

	// Публикуем в фоне
	go func() {
		ctxWithLogger := logger.WithContext(context.Background()).With(
			zap.String("project_id", projectUUID.String()),
			zap.String("user_id", userUUID.String()),
		)

		s.publishInBackground(context.Background(), target, userUUID, projectUUID, ctxWithLogger)
	}()

	return target, nil
}

// GetPublishStatus получает статус публикации
func (s *PublishService) GetPublishStatus(ctx context.Context, userID, targetID string) (*domain.PublishTarget, error) {
	targetUUID, err := uuid.Parse(targetID)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid target ID")
	}

	target, err := s.publishTargetRepo.GetByID(ctx, targetUUID.String())
	if err != nil {
		return nil, domain.ErrNotFound.WithMessage("target not found")
	}

	return target, nil
}

// GetPublishedURL получает URL опубликованного сайта
func (s *PublishService) GetPublishedURL(ctx context.Context, userID, targetID string) (string, error) {
	target, err := s.GetPublishStatus(ctx, userID, targetID)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", s.publicBaseURL(), target.Subdomain), nil
}

// PublishProject публикует проект
func (s *PublishService) PublishProject(ctx context.Context, userID, projectID uuid.UUID) (*PublishResult, error) {
	// Проверка доступа к проекту
	project, err := s.projectRepo.GetByID(ctx, projectID.String())
	if err != nil {
		return nil, domain.ErrNotFound.WithMessage("project not found")
	}

	if project.UserID != userID {
		return nil, domain.ErrForbidden.WithMessage("access denied")
	}

	// Проверяем наличие схемы
	if project.SchemaJSON == "" {
		return nil, domain.ErrBadRequest.WithMessage("project schema is empty")
	}

	// Генерируем уникальный subdomain на основе названия проекта
	subdomain := generateSubdomain(project.Name, projectID)

	// Создаём или обновляем цель публикации
	target := &domain.PublishTarget{
		ID:        uuid.New(),
		ProjectID: projectID,
		Subdomain: subdomain,
		Status:    domain.PublishStatusPublished,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Проверяем, есть ли уже цель публикации для этого проекта
	existingTarget, err := s.publishTargetRepo.GetByProjectID(ctx, projectID.String())
	if err == nil && existingTarget != nil {
		// Обновляем существующую цель
		target.ID = existingTarget.ID
		target.CreatedAt = existingTarget.CreatedAt
		if err := s.publishTargetRepo.Update(ctx, target); err != nil {
			return nil, domain.ErrInternal.WithError(err)
		}
	} else {
		// Создаём новую цель
		if err := s.publishTargetRepo.Create(ctx, target); err != nil {
			return nil, domain.ErrInternal.WithError(err)
		}
	}

	// Рендерим статический сайт
	buildDir, err := s.renderer.RenderStatic(ctx, projectID, project.SchemaJSON)
	if err != nil {
		return nil, domain.ErrInternal.WithMessage("failed to render static site")
	}

	// Загружаем файлы в S3/CDN
	remotePath := fmt.Sprintf("sites/%s", subdomain)
	if err := s.publisher.Upload(ctx, buildDir, remotePath); err != nil {
		return nil, domain.ErrInternal.WithMessage("failed to upload to storage")
	}

	now := time.Now()
	project.Status = domain.ProjectStatusPublished
	project.UpdatedAt = now
	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	publicURL := fmt.Sprintf("%s/sites/%s", s.publicBaseURL(), subdomain)

	return &PublishResult{
		Subdomain:   subdomain,
		PublicURL:   publicURL,
		PublishedAt: now.Format(time.RFC3339),
	}, nil
}

// UnpublishProject снимает проект с публикации
func (s *PublishService) UnpublishProject(ctx context.Context, userID, projectID uuid.UUID) error {
	project, err := s.projectRepo.GetByID(ctx, projectID.String())
	if err != nil {
		return domain.ErrNotFound.WithMessage("project not found")
	}

	if project.UserID != userID {
		return domain.ErrForbidden.WithMessage("access denied")
	}

	target, err := s.publishTargetRepo.GetByProjectID(ctx, projectID.String())
	if err != nil {
		return domain.ErrNotFound.WithMessage("publish target not found")
	}

	now := time.Now()

	target.Status = domain.PublishStatusDraft
	target.UpdatedAt = now
	if err := s.publishTargetRepo.Update(ctx, target); err != nil {
		return domain.ErrInternal.WithError(err)
	}

	project.Status = domain.ProjectStatusGenerated
	project.UpdatedAt = now
	if err := s.projectRepo.Update(ctx, project); err != nil {
		return domain.ErrInternal.WithError(err)
	}

	return nil
}

// ServePublished возвращает содержимое опубликованного проекта для указанного ресурса
func (s *PublishService) ServePublished(ctx context.Context, subdomain, assetPath string) (io.ReadCloser, string, error) {
	cleanPath := strings.TrimPrefix(assetPath, "/")
	if cleanPath == "" {
		cleanPath = "index.html"
	}
	cleanPath = filepath.Clean(cleanPath)
	if strings.Contains(cleanPath, "..") {
		return nil, "", domain.ErrForbidden
	}

	target, targetErr := s.publishTargetRepo.GetBySubdomain(ctx, subdomain)
	if targetErr != nil && !errors.Is(targetErr, domain.ErrNotFound) {
		return nil, "", targetErr
	}

	seen := make(map[string]struct{})
	var searchBases []string
	addBase := func(base string) {
		if _, ok := seen[base]; ok {
			return
		}
		seen[base] = struct{}{}
		searchBases = append(searchBases, base)
	}

	addBase(fmt.Sprintf("sites/%s", subdomain))

	if targetErr == nil && target != nil {
		if !strings.EqualFold(target.Subdomain, subdomain) {
			addBase(fmt.Sprintf("sites/%s", target.Subdomain))
		}
		addBase(fmt.Sprintf("sites/%s", target.ProjectID.String()))
	}

	for _, basePath := range searchBases {
		remotePath := filepath.Join(basePath, cleanPath)

		reader, contentType, err := s.publisher.GetObject(ctx, remotePath)
		if err == nil {
			return reader, contentType, nil
		}

		if !errors.Is(err, domain.ErrNotFound) {
			return nil, "", err
		}
	}

	return nil, "", domain.ErrNotFound
}

func (s *PublishService) publishInBackground(ctx context.Context, target *domain.PublishTarget, userID, projectID uuid.UUID, log logger.Logger) {
	log.Info("starting background publication")

	result, err := s.PublishProject(ctx, userID, projectID)
	if err != nil {
		log.Error("publication failed", zap.Error(err))
		target.Status = domain.PublishStatusFailed
	} else {
		log.Info("publication completed successfully",
			zap.String("public_url", result.PublicURL),
			zap.String("subdomain", result.Subdomain),
		)
		target.Status = domain.PublishStatusPublished
		now := time.Now()
		target.LastPublishedAt = &now
		target.UpdatedAt = now
	}

	if err := s.publishTargetRepo.Update(context.Background(), target); err != nil {
		log.Error("failed to update publish target", zap.Error(err))
	}
}
