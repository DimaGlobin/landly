package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/landly/backend/internal/logger"
	"github.com/landly/backend/internal/metrics"
	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/schema"
	"github.com/landly/backend/internal/storage/s3"
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
	DeletePrefix(ctx context.Context, prefix string) error
	// Новые методы для atomic publish и incremental upload
	UploadToRelease(ctx context.Context, localPath, basePath string, releaseID uuid.UUID) error
	CreateManifest(ctx context.Context, localPath string) (*s3.ReleaseManifest, error)
	UploadIncremental(ctx context.Context, localPath, basePath string, releaseID uuid.UUID, previousManifest *s3.ReleaseManifest) (*s3.ReleaseManifest, error)
	SetCurrentRelease(ctx context.Context, basePath string, releaseID uuid.UUID) error
	GetCurrentReleasePath(basePath string) string
	GetReleasePath(basePath string, releaseID uuid.UUID) string
	// CDN support
	HasCDN() bool
}

// PublishUserRepository интерфейс для получения пользователя
type PublishUserRepository interface {
	GetByID(ctx context.Context, userID string) (*domain.User, error)
}

// PublicBaseProvider интерфейс для получения базового URL приложения
// Позволяет динамически обновлять BASE_URL без перезапуска сервисов
type PublicBaseProvider interface {
	GetPublicBase() string
}

// StaticPublicBaseProvider статический провайдер базового URL
// Может быть обновлен через UpdateBase для изменения BASE_URL без перезапуска
type StaticPublicBaseProvider struct {
	base string
}

func NewStaticPublicBaseProvider(base string) *StaticPublicBaseProvider {
	return &StaticPublicBaseProvider{base: strings.TrimRight(base, "/")}
}

func (p *StaticPublicBaseProvider) GetPublicBase() string {
	if p.base != "" {
		return p.base
	}
	return "http://localhost:8080"
}

// UpdateBase обновляет базовый URL (для изменения BASE_URL без перезапуска)
func (p *StaticPublicBaseProvider) UpdateBase(newBase string) {
	p.base = strings.TrimRight(newBase, "/")
}

// PublishService сервис для публикации проектов
type PublishService struct {
	projectRepo        domain.ProjectRepository
	publishTargetRepo  domain.PublishTargetRepository
	publishJobRepo     domain.PublishJobRepository
	userRepo           PublishUserRepository
	renderer           Renderer
	publisher          Publisher
	publicBaseProvider PublicBaseProvider
}

// PublishResult результат публикации
type PublishResult struct {
	Subdomain   string `json:"subdomain"`
	Status      string `json:"status"`
	PublishedAt string `json:"published_at,omitempty"`
}

// NewPublishService создаёт новый publish service
func NewPublishService(
	projectRepo domain.ProjectRepository,
	publishTargetRepo domain.PublishTargetRepository,
	publishJobRepo domain.PublishJobRepository,
	userRepo PublishUserRepository,
	renderer Renderer,
	publisher Publisher,
	publicBaseProvider PublicBaseProvider,
) *PublishService {
	return &PublishService{
		projectRepo:        projectRepo,
		publishTargetRepo:  publishTargetRepo,
		publishJobRepo:     publishJobRepo,
		userRepo:           userRepo,
		renderer:           renderer,
		publisher:          publisher,
		publicBaseProvider: publicBaseProvider,
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
	if s.publicBaseProvider != nil {
		return s.publicBaseProvider.GetPublicBase()
	}
	return "http://localhost:8080"
}

// PublishSite публикует сайт (новый интерфейс)
func (s *PublishService) PublishSite(ctx context.Context, userID, projectID string, req *domain.PublishRequest) (*domain.PublishTarget, error) {
	log := logger.WithContext(ctx)
	log.Info("starting publish site", zap.String("user_id", userID), zap.String("project_id", projectID))

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		log.Error("invalid user ID", zap.String("user_id", userID), zap.Error(err))
		return nil, domain.ErrBadRequest.WithMessage("invalid user ID")
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		log.Error("invalid project ID", zap.String("project_id", projectID), zap.Error(err))
		return nil, domain.ErrBadRequest.WithMessage("invalid project ID")
	}

	// Проверка доступа к проекту
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		log.Error("project not found", zap.String("project_id", projectID), zap.Error(err))
		return nil, domain.ErrNotFound.WithMessage("project not found")
	}

	if project.UserID.String() != userID {
		log.Warn("access denied", zap.String("user_id", userID), zap.String("project_user_id", project.UserID.String()))
		return nil, domain.ErrForbidden.WithMessage("access denied")
	}

	// Создаем или обновляем цель публикации с использованием имени проекта
	// Формат: <base-url>/sites/<subdomain>
	subdomain := generateSubdomain(project.Name, projectUUID)
	var target *domain.PublishTarget

	existingTarget, err := s.publishTargetRepo.GetByProjectID(ctx, projectUUID.String())
	if err == nil && existingTarget != nil {
		log.Info("updating existing publish target", zap.String("target_id", existingTarget.ID.String()))
		existingTarget.Subdomain = subdomain
		existingTarget.Status = domain.PublishStatusDraft
		existingTarget.UpdatedAt = time.Now()
		if err := s.publishTargetRepo.Update(ctx, existingTarget); err != nil {
			log.Error("failed to update publish target", zap.Error(err))
			return nil, domain.ErrInternal.WithError(err)
		}
		target = existingTarget
	} else {
		log.Info("creating new publish target", zap.String("subdomain", subdomain))
		target = domain.NewPublishTarget(projectUUID, subdomain)
		if err := s.publishTargetRepo.Create(ctx, target); err != nil {
			log.Error("failed to create publish target", zap.Error(err))
			return nil, domain.ErrInternal.WithError(err)
		}
	}

	// Публикуем в фоне
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log := logger.WithContext(context.Background()).With(
					zap.String("project_id", projectUUID.String()),
					zap.String("user_id", userUUID.String()),
				)
				log.Error("panic in publishInBackground", zap.Any("panic", r))
			}
		}()

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
// Если CDN настроен, возвращает прямую ссылку на CDN, иначе - через наш сервер
func (s *PublishService) GetPublishedURL(ctx context.Context, userID, targetID string) (string, error) {
	target, err := s.GetPublishStatus(ctx, userID, targetID)
	if err != nil {
		return "", err
	}

	// Получаем URL от publisher (может быть CDN или S3)
	remotePath := fmt.Sprintf("sites/%s", target.Subdomain)
	directURL := s.publisher.GetPublicURL(remotePath)

	// Если GetPublicURL вернул URL, который не является нашим сервером,
	// значит это CDN или прямой S3 URL - используем его
	publicBase := s.publicBaseURL()
	if directURL != "" && !strings.HasPrefix(directURL, publicBase) {
		// Это CDN/S3 URL, возвращаем его напрямую
		return directURL, nil
	}

	// Иначе возвращаем URL через наш сервер (fallback для dev окружения)
	return fmt.Sprintf("%s/sites/%s", publicBase, target.Subdomain), nil
}

// PublishProject публикует проект с использованием atomic publish через версии
func (s *PublishService) PublishProject(ctx context.Context, userID, projectID uuid.UUID) (*PublishResult, error) {
	log := logger.WithContext(ctx)

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

	// Проверяем, нет ли уже активной задачи публикации
	pendingJob, err := s.publishJobRepo.GetPendingByProjectID(ctx, projectID.String())
	if err == nil && pendingJob != nil {
		return nil, domain.ErrBadRequest.WithMessage("publication already in progress")
	}

	// Генерируем уникальный subdomain на основе названия проекта
	subdomain := generateSubdomain(project.Name, projectID)

	// Получаем или создаем цель публикации
	target, err := s.publishTargetRepo.GetByProjectID(ctx, projectID.String())
	if err != nil {
		// Создаём новую цель
		target = &domain.PublishTarget{
			ID:        uuid.New(),
			ProjectID: projectID,
			Subdomain: subdomain,
			Status:    domain.PublishStatusDraft,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := s.publishTargetRepo.Create(ctx, target); err != nil {
			return nil, domain.ErrInternal.WithError(err)
		}
	} else {
		// Обновляем subdomain если изменился
		if target.Subdomain != subdomain {
			target.Subdomain = subdomain
			target.UpdatedAt = time.Now()
			if err := s.publishTargetRepo.Update(ctx, target); err != nil {
				return nil, domain.ErrInternal.WithError(err)
			}
		}
	}

	// Создаем задачу публикации
	releaseID := uuid.New()
	job := &domain.PublishJob{
		ID:              uuid.New(),
		ProjectID:       projectID,
		PublishTargetID: &target.ID,
		Status:          domain.PublishJobStatusRunning,
		ReleaseID:       &releaseID,
		StartedAt:       ptrTime(time.Now()),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	if err := s.publishJobRepo.Create(ctx, job); err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	// Нормализуем схему перед рендерингом
	normalizedSchema, err := schema.NormalizeSchema(project.SchemaJSON)
	if err != nil {
		log.Warn("failed to normalize schema, using original", zap.Error(err))
		normalizedSchema = project.SchemaJSON // Fallback на оригинальную схему
	}

	// Рендерим статический сайт
	buildDir, err := s.renderer.RenderStatic(ctx, projectID, normalizedSchema)
	if err != nil {
		job.Status = domain.PublishJobStatusFailed
		job.ErrorMessage = fmt.Sprintf("failed to render: %v", err)
		job.CompletedAt = ptrTime(time.Now())
		job.UpdatedAt = time.Now()
		if updateErr := s.publishJobRepo.Update(ctx, job); updateErr != nil {
			log.Warn("failed to update job status", zap.Error(updateErr))
		}
		return nil, domain.ErrInternal.WithMessage("failed to render static site")
	}

	// Получаем предыдущий manifest для incremental upload
	var previousManifest *s3.ReleaseManifest
	if target.CurrentReleaseID != nil {
		// Загружаем предыдущий manifest из S3
		basePath := fmt.Sprintf("sites/%s", subdomain)
		previousReleasePath := s.publisher.GetReleasePath(basePath, *target.CurrentReleaseID)
		manifestPath := filepath.Join(previousReleasePath, ".manifest.json")

		manifestReader, _, err := s.publisher.GetObject(ctx, manifestPath)
		if err == nil {
			defer manifestReader.Close()
			var manifest s3.ReleaseManifest
			manifestData, _ := io.ReadAll(manifestReader)
			if err := json.Unmarshal(manifestData, &manifest); err == nil {
				previousManifest = &manifest
			}
		}
	}

	// Загружаем файлы в release (incremental или полная загрузка)
	basePath := fmt.Sprintf("sites/%s", subdomain)
	var manifest *s3.ReleaseManifest
	if previousManifest != nil {
		// Incremental upload
		manifest, err = s.publisher.UploadIncremental(ctx, buildDir, basePath, releaseID, previousManifest)
		if err != nil {
			job.Status = domain.PublishJobStatusFailed
			job.ErrorMessage = fmt.Sprintf("failed to upload incrementally: %v", err)
			job.CompletedAt = ptrTime(time.Now())
			job.UpdatedAt = time.Now()
			if updateErr := s.publishJobRepo.Update(ctx, job); updateErr != nil {
				log.Warn("failed to update job status", zap.Error(updateErr))
			}
			return nil, domain.ErrInternal.WithMessage("failed to upload to storage")
		}
		// Manifest уже создан и сохранен в S3 через UploadIncremental
		if manifest != nil {
			log.Info("incremental upload completed",
				zap.String("release_id", releaseID.String()),
				zap.Int("files_count", len(manifest.Files)),
			)
		}
	} else {
		// Полная загрузка
		if err := s.publisher.UploadToRelease(ctx, buildDir, basePath, releaseID); err != nil {
			job.Status = domain.PublishJobStatusFailed
			job.ErrorMessage = fmt.Sprintf("failed to upload: %v", err)
			job.CompletedAt = ptrTime(time.Now())
			job.UpdatedAt = time.Now()
			if updateErr := s.publishJobRepo.Update(ctx, job); updateErr != nil {
				log.Warn("failed to update job status", zap.Error(updateErr))
			}
			return nil, domain.ErrInternal.WithMessage("failed to upload to storage")
		}
		// Создаем manifest после загрузки
		manifest, err = s.publisher.CreateManifest(ctx, buildDir)
		if err != nil {
			job.Status = domain.PublishJobStatusFailed
			job.ErrorMessage = fmt.Sprintf("failed to create manifest: %v", err)
			job.CompletedAt = ptrTime(time.Now())
			job.UpdatedAt = time.Now()
			if updateErr := s.publishJobRepo.Update(ctx, job); updateErr != nil {
				log.Warn("failed to update job status", zap.Error(updateErr))
			}
			return nil, domain.ErrInternal.WithMessage("failed to process files")
		}
		manifest.ReleaseID = releaseID
	}

	// Атомарно переключаем current на новый release
	if err := s.publisher.SetCurrentRelease(ctx, basePath, releaseID); err != nil {
		log.Error("failed to set current release", zap.Error(err), zap.String("release_id", releaseID.String()))
		job.Status = domain.PublishJobStatusFailed
		job.ErrorMessage = fmt.Sprintf("failed to switch release: %v", err)
		job.CompletedAt = ptrTime(time.Now())
		job.UpdatedAt = time.Now()
		if updateErr := s.publishJobRepo.Update(ctx, job); updateErr != nil {
			log.Warn("failed to update job status", zap.Error(updateErr))
		}
		return nil, domain.ErrInternal.WithMessage("failed to activate release")
	}

	// Обновляем target с новым release ID
	now := time.Now()
	target.CurrentReleaseID = &releaseID
	target.Status = domain.PublishStatusPublished
	target.LastPublishedAt = &now
	target.UpdatedAt = now
	if err := s.publishTargetRepo.Update(ctx, target); err != nil {
		log.Error("failed to update target", zap.Error(err))
		// Не возвращаем ошибку, т.к. файлы уже загружены
	}

	// Обновляем статус задачи
	job.Status = domain.PublishJobStatusSuccess
	job.CompletedAt = &now
	job.UpdatedAt = now
	if err := s.publishJobRepo.Update(ctx, job); err != nil {
		log.Warn("failed to update job status", zap.Error(err))
	}

	project.Status = domain.ProjectStatusPublished
	project.UpdatedAt = now
	if err := s.projectRepo.Update(ctx, project); err != nil {
		log.Warn("failed to update project status", zap.Error(err))
	}

	log.Info("publication completed successfully",
		zap.String("subdomain", subdomain),
		zap.String("release_id", releaseID.String()),
	)

	return &PublishResult{
		Subdomain:   subdomain,
		Status:      string(target.Status),
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

	// Record unpublish metric
	metrics.RecordPublish("unpublish")

	return nil
}

// ServePublished возвращает содержимое опубликованного проекта для указанного ресурса
func (s *PublishService) ServePublished(ctx context.Context, subdomain, assetPath string) (io.ReadCloser, string, error) {
	log := logger.WithContext(ctx).With(
		zap.String("subdomain", subdomain),
		zap.String("asset_path", assetPath),
	)

	log.Debug("ServePublished called",
		zap.String("subdomain", subdomain),
		zap.String("asset_path", assetPath))

	// Проверяем наличие publisher
	if s.publisher == nil {
		log.Error("publisher is nil")
		return nil, "", domain.ErrInternal.WithMessage("publisher not initialized")
	}

	cleanPath := strings.TrimPrefix(assetPath, "/")
	if cleanPath == "" {
		cleanPath = "index.html"
		log.Debug("empty asset path, defaulting to index.html")
	}

	// Reject path traversal before normalization
	if strings.Contains(assetPath, "..") {
		log.Warn("path traversal attempt detected", zap.String("asset_path", assetPath))
		return nil, "", domain.ErrForbidden
	}

	// Normalize path (use path package for URL-style forward slashes)
	cleanPath = path.Clean("/" + cleanPath)
	cleanPath = strings.TrimPrefix(cleanPath, "/")

	// Reject if Clean produced a path that escapes (starts with ..)
	if cleanPath == "" || strings.HasPrefix(cleanPath, "..") {
		log.Warn("invalid path after clean", zap.String("asset_path", assetPath), zap.String("clean_path", cleanPath))
		return nil, "", domain.ErrForbidden
	}

	log.Debug("looking up publish target by subdomain",
		zap.String("subdomain", subdomain))

	target, targetErr := s.publishTargetRepo.GetBySubdomain(ctx, subdomain)
	if targetErr != nil {
		if errors.Is(targetErr, domain.ErrNotFound) {
			log.Debug("publish target not found by subdomain",
				zap.String("subdomain", subdomain))
		} else {
			// Логируем внутреннюю ошибку с максимальным контекстом
			logger.LogInternalError(ctx, "error looking up publish target",
				targetErr,
				zap.String("subdomain", subdomain),
			)
			// Возвращаем общую ошибку без деталей
			return nil, "", domain.ErrInternal.WithMessage("failed to lookup publish target")
		}
	} else {
		log.Debug("publish target found",
			zap.String("target_id", target.ID.String()),
			zap.String("target_subdomain", target.Subdomain),
			zap.String("target_status", target.Status),
			zap.String("project_id", target.ProjectID.String()))
	}

	seen := make(map[string]struct{})
	var searchBases []string
	addBase := func(base string) {
		if _, ok := seen[base]; ok {
			log.Debug("skipping duplicate search base", zap.String("base", base))
			return
		}
		seen[base] = struct{}{}
		searchBases = append(searchBases, base)
		log.Debug("added search base", zap.String("base", base))
	}

	// Приоритет 1: current симлинк (для atomic publish)
	basePath := fmt.Sprintf("sites/%s", subdomain)
	currentPath := s.publisher.GetCurrentReleasePath(basePath)
	log.Debug("generating search paths",
		zap.String("base_path", basePath),
		zap.String("current_release_path", currentPath))

	addBase(currentPath)

	// Приоритет 2: прямой путь (для обратной совместимости)
	addBase(basePath)

	if targetErr == nil && target != nil {
		if !strings.EqualFold(target.Subdomain, subdomain) {
			// Если subdomain изменился, проверяем оба
			altBasePath := fmt.Sprintf("sites/%s", target.Subdomain)
			altCurrentPath := s.publisher.GetCurrentReleasePath(altBasePath)
			log.Debug("subdomain mismatch, adding alternative paths",
				zap.String("requested_subdomain", subdomain),
				zap.String("target_subdomain", target.Subdomain),
				zap.String("alt_base_path", altBasePath),
				zap.String("alt_current_path", altCurrentPath))
			addBase(altCurrentPath)
			addBase(altBasePath)
		}
		// Fallback на project ID
		projectBasePath := fmt.Sprintf("sites/%s", target.ProjectID.String())
		projectCurrentPath := s.publisher.GetCurrentReleasePath(projectBasePath)
		log.Debug("adding project ID fallback paths",
			zap.String("project_id", target.ProjectID.String()),
			zap.String("project_base_path", projectBasePath),
			zap.String("project_current_path", projectCurrentPath))
		addBase(projectCurrentPath)
		addBase(projectBasePath)
	}

	log.Info("searching for file",
		zap.String("subdomain", subdomain),
		zap.String("asset_path", assetPath),
		zap.String("clean_path", cleanPath),
		zap.Int("search_bases_count", len(searchBases)),
		zap.Strings("search_bases", searchBases))

	for i, searchBase := range searchBases {
		// Используем прямые слэши для S3-совместимых путей
		searchBase = strings.TrimSuffix(searchBase, "/")
		// Не модифицируем cleanPath здесь, так как он используется в следующей итерации
		cleanPathForSearch := strings.TrimPrefix(cleanPath, "/")
		remotePath := searchBase + "/" + cleanPathForSearch

		log.Debug("trying to get object from S3",
			zap.Int("attempt", i+1),
			zap.Int("total_attempts", len(searchBases)),
			zap.String("search_base", searchBase),
			zap.String("clean_path_for_search", cleanPathForSearch),
			zap.String("remote_path", remotePath))

		reader, contentType, err := s.publisher.GetObject(ctx, remotePath)
		if err == nil {
			log.Info("file found in S3",
				zap.String("remote_path", remotePath),
				zap.String("content_type", contentType),
				zap.Int("attempt", i+1))
			return reader, contentType, nil
		}

		if !errors.Is(err, domain.ErrNotFound) {
			log.Error("error getting object from S3",
				zap.String("remote_path", remotePath),
				zap.String("error_type", fmt.Sprintf("%T", err)),
				zap.Error(err))
			return nil, "", err
		}

		log.Debug("file not found at path",
			zap.String("remote_path", remotePath),
			zap.Int("attempt", i+1))
	}

	log.Warn("file not found in any search path",
		zap.String("subdomain", subdomain),
		zap.String("asset_path", assetPath),
		zap.String("clean_path", cleanPath),
		zap.Int("search_bases_count", len(searchBases)),
		zap.Strings("search_bases", searchBases))

	return nil, "", domain.ErrNotFound
}

// HasCDN проверяет, настроен ли CDN
func (s *PublishService) HasCDN() bool {
	if s == nil {
		logger.WithContext(context.Background()).Error("PublishService is nil in HasCDN")
		return false
	}
	if s.publisher == nil {
		logger.WithContext(context.Background()).Error("publisher is nil in HasCDN")
		return false
	}
	hasCDN := s.publisher.HasCDN()
	logger.WithContext(context.Background()).Debug("HasCDN check",
		zap.Bool("has_cdn", hasCDN))
	return hasCDN
}

// GetCDNURL возвращает CDN URL для указанного пути
func (s *PublishService) GetCDNURL(remotePath string) string {
	if s == nil {
		logger.WithContext(context.Background()).Error("PublishService is nil in GetCDNURL",
			zap.String("remote_path", remotePath))
		return ""
	}
	if s.publisher == nil {
		logger.WithContext(context.Background()).Error("publisher is nil in GetCDNURL",
			zap.String("remote_path", remotePath))
		return ""
	}
	cdnURL := s.publisher.GetPublicURL(remotePath)
	logger.WithContext(context.Background()).Debug("GetCDNURL",
		zap.String("remote_path", remotePath),
		zap.String("cdn_url", cdnURL))
	return cdnURL
}

func (s *PublishService) publishInBackground(ctx context.Context, target *domain.PublishTarget, userID, projectID uuid.UUID, log logger.Logger) {
	log.Info("starting background publication")

	result, err := s.PublishProject(ctx, userID, projectID)
	if err != nil {
		log.Error("publication failed", zap.Error(err))
		target.Status = domain.PublishStatusFailed
	} else {
		log.Info("publication completed successfully",
			zap.String("subdomain", result.Subdomain),
			zap.String("status", result.Status),
		)
		target.Status = domain.PublishStatusPublished
		now := time.Now()
		target.LastPublishedAt = &now
		target.UpdatedAt = now
		// Record publish metric
		metrics.RecordPublish("publish")
	}

	if err := s.publishTargetRepo.Update(context.Background(), target); err != nil {
		log.Error("failed to update publish target", zap.Error(err))
	}
}
