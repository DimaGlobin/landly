package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/landly/backend/internal/logger"
	"github.com/landly/backend/internal/schema"
	domain "github.com/landly/backend/internal/models"
	"go.uber.org/zap"
)

// SimpleGenerateService простой сервис генерации
type SimpleGenerateService struct {
	projectRepo domain.ProjectRepository
	versionRepo domain.SchemaVersionRepository
	aiClient    AIClient
}

// NewSimpleGenerateService создает новый простой сервис генерации
func NewSimpleGenerateService(projectRepo domain.ProjectRepository, versionRepo domain.SchemaVersionRepository, aiClient AIClient) *SimpleGenerateService {
	return &SimpleGenerateService{
		projectRepo: projectRepo,
		versionRepo: versionRepo,
		aiClient:    aiClient,
	}
}

// GenerateSimple простая генерация лендинга
func (s *SimpleGenerateService) GenerateSimple(ctx context.Context, userID, projectID string, prompt, paymentURL string) (map[string]interface{}, error) {
	log := logger.WithContext(ctx).With(
		zap.String("project_id", projectID),
		zap.String("user_id", userID),
	)

	log.Info("starting simple generation",
		zap.String("prompt", prompt),
	)

	// Проверяем доступ к проекту
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("проект не найден: %w", err)
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, fmt.Errorf("неверный ID проекта: %w", err)
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("неверный ID пользователя: %w", err)
	}

	if project.UserID != userUUID {
		return nil, fmt.Errorf("доступ запрещен")
	}

	// Проверяем, что проект принадлежит пользователю
	_ = projectUUID // Используем переменную для избежания ошибки компиляции

	// Генерируем схему с помощью AI
	startTime := time.Now()
	log.Info("generating schema with AI")
	schemaJSON, err := s.aiClient.GenerateLandingSchema(ctx, prompt, paymentURL)
	duration := time.Since(startTime)
	if err != nil {
		log.Error("AI generation failed",
			zap.Error(err),
			zap.Duration("duration_ms", duration),
		)
		return nil, fmt.Errorf("ошибка AI генерации: %w", err)
	}

	log.Info("AI generated schema successfully",
		zap.Int("schema_length", len(schemaJSON)),
		zap.Duration("duration_ms", duration),
		zap.Int("tokens_used", 0), // Mocked AI, no tokens
	)

	// Валидируем схему перед сохранением
	if err := schema.ValidateSchema(schemaJSON); err != nil {
		log.Error("schema validation failed",
			zap.Error(err),
			zap.Duration("duration_ms", duration),
		)
		return nil, fmt.Errorf("невалидная схема: %w", err)
	}

	// Сохраняем текущую схему как версию перед обновлением
	if project.SchemaJSON != "" && s.versionRepo != nil {
		currentVersion := domain.NewSchemaVersion(
			projectUUID,
			userUUID,
			project.SchemaJSON,
			domain.SchemaVersionSourceGenerate,
		)
		if err := s.versionRepo.Create(ctx, currentVersion); err != nil {
			log.Warn("failed to save current schema as version", zap.Error(err))
			// Продолжаем, даже если не удалось сохранить версию
		}
	}

	// Парсим JSON схему
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return nil, fmt.Errorf("ошибка парсинга схемы: %w", err)
	}

	// Сохраняем схему в проект
	log.Info("saving schema to project")
	if err := s.projectRepo.UpdateSchema(ctx, projectID, schemaJSON); err != nil {
		log.Error("failed to save schema",
			zap.Error(err),
			zap.Duration("duration_ms", duration),
		)
		return nil, fmt.Errorf("ошибка сохранения схемы: %w", err)
	}

	// Сохраняем новую версию (если versionRepo доступен)
	if s.versionRepo != nil {
		newVersion := domain.NewSchemaVersion(
			projectUUID,
			userUUID,
			schemaJSON,
			domain.SchemaVersionSourceGenerate,
		)
		if err := s.versionRepo.Create(ctx, newVersion); err != nil {
			log.Warn("failed to save new schema version", zap.Error(err))
			// Продолжаем, даже если не удалось сохранить версию
		}
	}

	log.Info("schema saved to project successfully",
		zap.Duration("duration_ms", duration),
	)

	return schema, nil
}
