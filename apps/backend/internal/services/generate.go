package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/landly/backend/internal/logger"
	domain "github.com/landly/backend/internal/models"
	"go.uber.org/zap"
)

// AIClient интерфейс для AI-генерации
type AIClient interface {
	GenerateLandingSchema(ctx context.Context, prompt, paymentURL string) (string, error)
}

// GenerateService сервис для генерации лендингов
type GenerateService struct {
	projectRepo     domain.ProjectRepository
	integrationRepo domain.IntegrationRepository
	sessionRepo     domain.GenerationSessionRepository
	aiClient        AIClient
}

// NewGenerateService создаёт новый generate service
func NewGenerateService(
	projectRepo domain.ProjectRepository,
	integrationRepo domain.IntegrationRepository,
	sessionRepo domain.GenerationSessionRepository,
	aiClient AIClient,
) *GenerateService {
	return &GenerateService{
		projectRepo:     projectRepo,
		integrationRepo: integrationRepo,
		sessionRepo:     sessionRepo,
		aiClient:        aiClient,
	}
}

// GenerateSite генерирует лендинг (новый интерфейс)
func (s *GenerateService) GenerateSite(ctx context.Context, userID, projectID string, req *domain.GenerateRequest) (*domain.GenerationSession, error) {
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid user ID")
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid project ID")
	}

	// Создаем сессию генерации
	session := domain.NewGenerationSession(projectUUID, req.Prompt, "gpt-4")
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	defer func() {
		if err := s.sessionRepo.Update(context.Background(), session); err != nil {
			log := logger.WithContext(ctx).With(
				zap.String("project_id", projectUUID.String()),
				zap.String("user_id", userUUID.String()),
			)
			log.Error("failed to update session status", zap.Error(err))
		}
	}()

	// Запускаем генерацию синхронно
	log := logger.WithContext(ctx).With(
		zap.String("project_id", projectUUID.String()),
		zap.String("user_id", userUUID.String()),
	)

	log.Info("starting landing page generation",
		zap.String("prompt", req.Prompt),
		zap.String("payment_url", req.PaymentURL),
	)

	updatedProject, err := s.GenerateLanding(context.Background(), userUUID, projectUUID, req.Prompt, req.PaymentURL)
	if err != nil {
		log.Error("generation failed", zap.Error(err))
		session.Status = domain.GenerationStatusFailed
		session.CompletedAt = ptrTime(time.Now())
		return session, domain.ErrInternal.WithError(err)
	}

	log.Info("generation completed successfully")
	session.Status = domain.GenerationStatusCompleted
	session.CompletedAt = ptrTime(time.Now())
	if updatedProject != nil {
		session.SchemaJSON = updatedProject.SchemaJSON
	}

	// Обновляем сессию
	if updateErr := s.sessionRepo.Update(context.Background(), session); updateErr != nil {
		log.Error("failed to update session", zap.Error(updateErr))
	} else {
		log.Info("session updated in database")
	}

	return session, nil
}

// GetGenerationStatus получает статус генерации
func (s *GenerateService) GetGenerationStatus(ctx context.Context, userID, sessionID string) (*domain.GenerationSession, error) {
	sessionUUID, err := uuid.Parse(sessionID)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid session ID")
	}

	session, err := s.sessionRepo.GetByID(ctx, sessionUUID.String())
	if err != nil {
		return nil, domain.ErrNotFound.WithMessage("session not found")
	}

	return session, nil
}

// GetGenerationResult получает результат генерации
func (s *GenerateService) GetGenerationResult(ctx context.Context, userID, sessionID string) (*domain.GenerationResult, error) {
	session, err := s.GetGenerationStatus(ctx, userID, sessionID)
	if err != nil {
		return nil, err
	}

	if session.Status != domain.GenerationStatusCompleted {
		return nil, domain.ErrBadRequest.WithMessage("generation not completed")
	}

	var schema domain.PageSchema
	if err := json.Unmarshal([]byte(session.SchemaJSON), &schema); err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	return &domain.GenerationResult{
		Schema: &schema,
		HTML:   "", // TODO: генерировать HTML
	}, nil
}

// GenerateLanding генерирует лендинг с помощью AI
func (s *GenerateService) GenerateLanding(ctx context.Context, userID, projectID uuid.UUID, prompt, paymentURL string) (*domain.Project, error) {
	// Проверка доступа к проекту
	project, err := s.projectRepo.GetByID(ctx, projectID.String())
	if err != nil {
		return nil, domain.ErrNotFound.WithMessage("project not found")
	}

	if project.UserID != userID {
		return nil, domain.ErrForbidden.WithMessage("access denied")
	}

	// Создаём сессию генерации
	session := &domain.GenerationSession{
		ID:        uuid.New(),
		ProjectID: projectID,
		Prompt:    prompt,
		Status:    domain.GenerationStatusPending,
		CreatedAt: time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	defer func() {
		if err := s.sessionRepo.Update(context.Background(), session); err != nil {
			log := logger.WithContext(ctx).With(
				zap.String("project_id", projectID.String()),
				zap.String("user_id", userID.String()),
			)
			log.Error("failed to update session status", zap.Error(err))
		}
	}()

	// Генерируем схему с помощью AI
	log := logger.WithContext(ctx).With(
		zap.String("project_id", projectID.String()),
		zap.String("user_id", userID.String()),
	)

	log.Info("calling AI client for schema generation")
	schemaJSON, err := s.aiClient.GenerateLandingSchema(ctx, prompt, paymentURL)
	if err != nil {
		log.Error("AI generation failed", zap.Error(err))
		// Обновляем статус сессии на ошибку
		session.Status = domain.GenerationStatusFailed
		return nil, domain.ErrInternal.WithMessage("AI generation failed")
	}

	log.Info("AI generated schema successfully",
		zap.Int("schema_length", len(schemaJSON)),
	)

	// Сохраняем схему в проект
	log.Info("saving schema to project")
	if err := s.projectRepo.UpdateSchema(ctx, projectID.String(), schemaJSON); err != nil {
		log.Error("failed to save schema to project", zap.Error(err))
		session.Status = domain.GenerationStatusFailed
		return nil, domain.ErrInternal.WithError(err)
	}
	log.Info("schema saved to project successfully")

	session.Status = domain.GenerationStatusCompleted
	session.SchemaJSON = schemaJSON
	session.CompletedAt = ptrTime(time.Now())

	// Получаем обновлённый проект
	updatedProject, err := s.projectRepo.GetByID(ctx, projectID.String())
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	return updatedProject, nil
}

// GetPreview получает превью проекта
func (s *GenerateService) GetPreview(ctx context.Context, userID, projectID uuid.UUID) (map[string]interface{}, error) {
	// Проверка доступа к проекту
	project, err := s.projectRepo.GetByID(ctx, projectID.String())
	if err != nil {
		return nil, domain.ErrNotFound.WithMessage("project not found")
	}

	if project.UserID != userID {
		return nil, domain.ErrForbidden.WithMessage("access denied")
	}

	// Парсим схему
	var schema map[string]interface{}
	if project.SchemaJSON != "" {
		if err := json.Unmarshal([]byte(project.SchemaJSON), &schema); err != nil {
			return nil, domain.ErrInternal.WithMessage("invalid schema format")
		}
	}

	return schema, nil
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
