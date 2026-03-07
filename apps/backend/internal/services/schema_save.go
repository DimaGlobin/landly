package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/landly/backend/internal/logger"
	domain "github.com/landly/backend/internal/models"
	"go.uber.org/zap"
)

// saveSchemaWithVersioning saves a new schema for a project, recording before/after versions.
// Steps: save current schema as version → UpdateSchema → save new schema as version.
// Version failures are non-fatal (warn-and-continue); UpdateSchema failure is returned.
func saveSchemaWithVersioning(
	ctx context.Context,
	projectRepo domain.ProjectRepository,
	versionRepo domain.SchemaVersionRepository,
	project *domain.Project,
	userID uuid.UUID,
	newSchemaJSON string,
	source domain.SchemaVersionSource,
	log logger.Logger,
) error {
	if project.SchemaJSON != "" && versionRepo != nil {
		current := domain.NewSchemaVersion(project.ID, userID, project.SchemaJSON, source)
		if err := versionRepo.Create(ctx, current); err != nil {
			log.Warn("failed to save current schema as version", zap.Error(err))
		}
	}

	if err := projectRepo.UpdateSchema(ctx, project.ID.String(), newSchemaJSON); err != nil {
		return err
	}

	if versionRepo != nil {
		newVer := domain.NewSchemaVersion(project.ID, userID, newSchemaJSON, source)
		if err := versionRepo.Create(ctx, newVer); err != nil {
			log.Warn("failed to save new schema version", zap.Error(err))
		}
	}

	return nil
}
