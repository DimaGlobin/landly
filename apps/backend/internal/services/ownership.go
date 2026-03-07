package services

import (
	"context"

	domain "github.com/landly/backend/internal/models"
)

// ensureProjectOwnership fetches a project and verifies the given user owns it.
func ensureProjectOwnership(ctx context.Context, projectRepo domain.ProjectRepository, userID, projectID string) (*domain.Project, error) {
	project, err := projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, domain.ErrNotFound.WithMessage("project not found")
	}

	if project.UserID.String() != userID {
		return nil, domain.ErrForbidden.WithMessage("access denied")
	}

	return project, nil
}
