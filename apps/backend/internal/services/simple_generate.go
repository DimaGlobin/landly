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

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid user ID")
	}

	project, err := ensureProjectOwnership(ctx, s.projectRepo, userID, projectID)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	log.Info("generating schema with AI")
	schemaJSON, err := s.aiClient.GenerateLandingSchema(ctx, prompt, paymentURL)
	duration := time.Since(startTime)
	if err != nil {
		log.Error("AI generation failed",
			zap.Error(err),
			zap.Duration("duration_ms", duration),
		)
		return nil, domain.ErrInternal.WithError(err)
	}

	log.Info("AI generated schema successfully",
		zap.Int("schema_length", len(schemaJSON)),
		zap.Duration("duration_ms", duration),
	)

	if err := schema.ValidateSchema(schemaJSON); err != nil {
		log.Error("schema validation failed",
			zap.Error(err),
			zap.Duration("duration_ms", duration),
		)
		return nil, domain.ErrBadRequest.WithMessage(fmt.Sprintf("invalid schema: %v", err))
	}

	log.Info("saving schema to project")
	if err := saveSchemaWithVersioning(ctx, s.projectRepo, s.versionRepo, project, userUUID, schemaJSON, domain.SchemaVersionSourceGenerate, log); err != nil {
		log.Error("failed to save schema",
			zap.Error(err),
			zap.Duration("duration_ms", duration),
		)
		return nil, domain.ErrInternal.WithError(err)
	}

	log.Info("schema saved to project successfully",
		zap.Duration("duration_ms", duration),
	)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(schemaJSON), &result); err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	return result, nil
}
