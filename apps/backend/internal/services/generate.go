package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	messageRepo     domain.GenerationMessageRepository
	aiClient        AIClient
}

// NewGenerateService создаёт новый generate service
func NewGenerateService(
	projectRepo domain.ProjectRepository,
	integrationRepo domain.IntegrationRepository,
	sessionRepo domain.GenerationSessionRepository,
	messageRepo domain.GenerationMessageRepository,
	aiClient AIClient,
) *GenerateService {
	return &GenerateService{
		projectRepo:     projectRepo,
		integrationRepo: integrationRepo,
		sessionRepo:     sessionRepo,
		messageRepo:     messageRepo,
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

	normalizedSchema, autoFixes, err := sanitizeSchema(schemaJSON)
	if err != nil {
		log.Error("schema validation failed", zap.Error(err))
		session.Status = domain.GenerationStatusFailed
		return nil, domain.ErrInternal.WithMessage("generated schema is invalid")
	}
	if len(autoFixes) > 0 {
		log.Info("schema normalized", zap.Strings("auto_fixes", autoFixes))
	}
	schemaJSON = normalizedSchema

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

// GetChatHistory возвращает текущую сессию и историю сообщений для проекта
func (s *GenerateService) GetChatHistory(ctx context.Context, userID, projectID string) (*domain.GenerationSession, []*domain.GenerationMessage, error) {
	project, err := s.ensureProjectOwnership(ctx, userID, projectID)
	if err != nil {
		return nil, nil, err
	}

	session, err := s.ensureSessionForProject(ctx, project)
	if err != nil {
		return nil, nil, err
	}

	messages, err := s.messageRepo.ListBySession(ctx, session.ID.String())
	if err != nil {
		return nil, nil, err
	}

	return session, messages, nil
}

// SendChatMessage обрабатывает новое сообщение пользователя и возвращает обновлённую историю
func (s *GenerateService) SendChatMessage(ctx context.Context, userID, projectID, content string) (*domain.GenerationSession, []*domain.GenerationMessage, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return nil, nil, domain.ErrBadRequest.WithMessage("message content is required")
	}

	project, err := s.ensureProjectOwnership(ctx, userID, projectID)
	if err != nil {
		return nil, nil, err
	}

	session, err := s.ensureSessionForProject(ctx, project)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()
	userMessage := &domain.GenerationMessage{
		ID:         uuid.New(),
		SessionID:  session.ID,
		Role:       domain.MessageRoleUser,
		Content:    trimmed,
		Metadata:   "",
		TokensUsed: 0,
		CreatedAt:  now,
	}

	if err := s.messageRepo.Create(ctx, userMessage); err != nil {
		return nil, nil, err
	}

	messages, err := s.messageRepo.ListBySession(ctx, session.ID.String())
	if err != nil {
		return nil, nil, err
	}

	prompt := s.buildChatPrompt(session, messages)

	log := logger.WithContext(ctx).With(
		zap.String("project_id", project.ID.String()),
		zap.String("user_id", project.UserID.String()),
	)

	log.Info("generating landing schema via chat", zap.String("prompt_snippet", truncateForLog(prompt)))

	schemaJSON, err := s.aiClient.GenerateLandingSchema(ctx, prompt, "")
	if err != nil {
		log.Error("chat generation failed", zap.Error(err))
		session.Status = domain.GenerationStatusFailed
		session.UpdatedAt = now
		session.CompletedAt = ptrTime(now)
		_ = s.sessionRepo.Update(ctx, session)
		return nil, nil, domain.ErrInternal.WithError(err)
	}

	normalizedSchema, autoFixes, err := sanitizeSchema(schemaJSON)
	if err != nil {
		log.Error("chat schema validation failed", zap.Error(err))
		return nil, nil, domain.ErrInternal.WithMessage("generated schema is invalid")
	}
	if len(autoFixes) > 0 {
		log.Info("schema normalized", zap.Strings("auto_fixes", autoFixes))
	}
	schemaJSON = normalizedSchema

	if err := s.projectRepo.UpdateSchema(ctx, project.ID.String(), schemaJSON); err != nil {
		log.Error("failed to persist generated schema", zap.Error(err))
		return nil, nil, err
	}

	project.SchemaJSON = schemaJSON
	project.Status = domain.ProjectStatusGenerated
	project.UpdatedAt = now

	session.Prompt = trimmed
	session.Status = domain.GenerationStatusCompleted
	session.SchemaJSON = schemaJSON
	session.CompletedAt = ptrTime(now)
	session.UpdatedAt = now
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		log.Error("failed to update generation session", zap.Error(err))
		return nil, nil, err
	}

	assistantContent := fmt.Sprintf("Готово! Я обновил лендинг согласно запросу: \"%s\". Посмотри предпросмотр справа.", truncateUserContent(trimmed))
	assistantMessage := &domain.GenerationMessage{
		ID:         uuid.New(),
		SessionID:  session.ID,
		Role:       domain.MessageRoleAssistant,
		Content:    assistantContent,
		Metadata:   fmt.Sprintf("{\"schema_updated\":true,\"schema_length\":%d}", len(schemaJSON)),
		TokensUsed: 0,
		CreatedAt:  time.Now(),
	}

	if err := s.messageRepo.Create(ctx, assistantMessage); err != nil {
		log.Error("failed to save assistant message", zap.Error(err))
		return nil, nil, err
	}

	messages = append(messages, assistantMessage)

	log.Info("chat generation completed successfully",
		zap.Int("messages_total", len(messages)),
	)

	return session, messages, nil
}

func (s *GenerateService) ensureProjectOwnership(ctx context.Context, userID, projectID string) (*domain.Project, error) {
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, domain.ErrNotFound.WithMessage("project not found")
	}

	if project.UserID.String() != userID {
		return nil, domain.ErrForbidden.WithMessage("access denied")
	}

	return project, nil
}

func (s *GenerateService) ensureSessionForProject(ctx context.Context, project *domain.Project) (*domain.GenerationSession, error) {
	sessions, err := s.sessionRepo.GetByProjectID(ctx, project.ID.String())
	if err != nil {
		return nil, err
	}

	if len(sessions) > 0 {
		session := sessions[0]
		if session.SchemaJSON == "" {
			session.SchemaJSON = project.SchemaJSON
		}
		return session, nil
	}

	session := domain.NewGenerationSession(project.ID, "", "gpt-4")
	session.SchemaJSON = project.SchemaJSON
	session.CreatedAt = time.Now()
	session.UpdatedAt = session.CreatedAt

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	return session, nil
}

func (s *GenerateService) buildChatPrompt(session *domain.GenerationSession, messages []*domain.GenerationMessage) string {
	var builder strings.Builder

	if session.SchemaJSON != "" {
		builder.WriteString("Текущая схема лендинга (JSON):\n")
		builder.WriteString(session.SchemaJSON)
		builder.WriteString("\n\n")
	}

	builder.WriteString("История диалога:\n")
	for _, msg := range messages {
		label := "Система"
		switch msg.Role {
		case domain.MessageRoleUser:
			label = "Пользователь"
		case domain.MessageRoleAssistant:
			label = "Ассистент"
		}
		builder.WriteString(label)
		builder.WriteString(": ")
		builder.WriteString(msg.Content)
		builder.WriteString("\n\n")
	}

	builder.WriteString("На основе истории обнови JSON-схему лендинга. Верни только JSON без дополнительного текста.")

	return builder.String()
}

func truncateUserContent(content string) string {
	const limit = 120
	runes := []rune(content)
	if len(runes) <= limit {
		return content
	}
	return string(runes[:limit]) + "..."
}

func truncateForLog(content string) string {
	const limit = 200
	runes := []rune(content)
	if len(runes) <= limit {
		return content
	}
	return string(runes[:limit]) + "..."
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
