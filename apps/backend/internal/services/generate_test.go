package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/services/mocks"
)

func TestGenerateService_GenerateLanding_Success(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	projectRepo := new(mocks.ProjectRepositoryMock)
	sessionRepo := new(mocks.GenerationSessionRepositoryMock)
	aiClient := new(mocks.AIClientMock)

	svc := NewGenerateService(projectRepo, nil, sessionRepo, aiClient)

	project := &domain.Project{ID: projectID, UserID: userID}
	projectRepo.On("GetByID", ctx, projectID.String()).Return(project, nil).Once()
	sessionRepo.On("Create", ctx, mock.MatchedBy(func(session *domain.GenerationSession) bool {
		require.Equal(t, projectID, session.ProjectID)
		require.Equal(t, "Prompt", session.Prompt)
		return true
	})).Return(nil).Once()
	generatedSchema := `{"pages":[{"path":"/","title":"Home","blocks":[]}]} `
	aiClient.On("GenerateLandingSchema", ctx, "Prompt", "https://pay").Return(generatedSchema, nil).Once()
	projectRepo.On("UpdateSchema", ctx, projectID.String(), generatedSchema).Return(nil).Once()
	projectRepo.On("GetByID", ctx, projectID.String()).Return(&domain.Project{ID: projectID, UserID: userID, SchemaJSON: generatedSchema}, nil).Once()
	sessionRepo.On("Update", mock.Anything, mock.MatchedBy(func(session *domain.GenerationSession) bool {
		return session.Status == domain.GenerationStatusCompleted
	})).Return(nil)

	updated, err := svc.GenerateLanding(ctx, userID, projectID, "Prompt", "https://pay")
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, generatedSchema, updated.SchemaJSON)

	projectRepo.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
	aiClient.AssertExpectations(t)
}

func TestGenerateService_GenerateLanding_AIError(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	projectRepo := new(mocks.ProjectRepositoryMock)
	sessionRepo := new(mocks.GenerationSessionRepositoryMock)
	aiClient := new(mocks.AIClientMock)

	svc := NewGenerateService(projectRepo, nil, sessionRepo, aiClient)

	project := &domain.Project{ID: projectID, UserID: userID}
	projectRepo.On("GetByID", ctx, projectID.String()).Return(project, nil)
	sessionRepo.On("Create", ctx, mock.MatchedBy(func(session *domain.GenerationSession) bool {
		return session.ProjectID == projectID
	})).Return(nil)
	aiClient.On("GenerateLandingSchema", ctx, "Prompt", "https://pay").Return("", errors.New("ai down"))
	sessionRepo.On("Update", mock.Anything, mock.MatchedBy(func(session *domain.GenerationSession) bool {
		return session.Status == domain.GenerationStatusFailed
	})).Return(nil)

	updated, err := svc.GenerateLanding(ctx, userID, projectID, "Prompt", "https://pay")
	assert.Nil(t, updated)
	assert.Error(t, err)

	projectRepo.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
	aiClient.AssertExpectations(t)
}

func TestGenerateService_GenerateSite_InvalidUser(t *testing.T) {
	ctx := context.Background()
	projectRepo := new(mocks.ProjectRepositoryMock)
	sessionRepo := new(mocks.GenerationSessionRepositoryMock)
	aiClient := new(mocks.AIClientMock)

	svc := NewGenerateService(projectRepo, nil, sessionRepo, aiClient)

	sessionRepo.On("Create", ctx, mock.AnythingOfType("*models.GenerationSession")).Return(nil)
	sessionRepo.On("Update", mock.Anything, mock.AnythingOfType("*models.GenerationSession")).Return(nil)
	aiClient.On("GenerateLandingSchema", mock.Anything, mock.Anything, mock.Anything).Return("{}", nil)
	projectRepo.On("UpdateSchema", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	projectRepo.On("GetByID", mock.Anything, mock.Anything).Return(&domain.Project{UserID: uuid.New()}, nil)

	session, err := svc.GenerateSite(ctx, "bad-user", uuid.New().String(), &domain.GenerateRequest{Prompt: "p"})
	assert.Nil(t, session)
	assert.Error(t, err)
}
