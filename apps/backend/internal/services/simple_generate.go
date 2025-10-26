package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/landly/backend/internal/logger"
	domain "github.com/landly/backend/internal/models"
	"go.uber.org/zap"
)

// SimpleGenerateService простой сервис генерации
type SimpleGenerateService struct {
	projectRepo domain.ProjectRepository
	aiClient    AIClient
}

// NewSimpleGenerateService создает новый простой сервис генерации
func NewSimpleGenerateService(projectRepo domain.ProjectRepository, aiClient AIClient) *SimpleGenerateService {
	return &SimpleGenerateService{
		projectRepo: projectRepo,
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
	log.Info("generating schema with AI")
	schemaJSON, err := s.aiClient.GenerateLandingSchema(ctx, prompt, paymentURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка AI генерации: %w", err)
	}

	log.Info("AI generated schema successfully",
		zap.Int("schema_length", len(schemaJSON)),
	)

	// Парсим JSON схему
	var schema map[string]interface{}
	if err := json.Unmarshal([]byte(schemaJSON), &schema); err != nil {
		return nil, fmt.Errorf("ошибка парсинга схемы: %w", err)
	}

	// Сохраняем схему в проект
	log.Info("saving schema to project")
	if err := s.projectRepo.UpdateSchema(ctx, projectID, schemaJSON); err != nil {
		return nil, fmt.Errorf("ошибка сохранения схемы: %w", err)
	}

	log.Info("schema saved to project successfully")

	return schema, nil
}
