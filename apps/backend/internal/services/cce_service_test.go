package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/services/mocks"
)

func TestCCEServiceBuildPromptContext_Defaults(t *testing.T) {
	projectRepo := new(mocks.ProjectRepositoryMock)
	brandRepo := new(mocks.BrandProfileRepositoryMock)
	productRepo := new(mocks.ProductProfileRepositoryMock)
	snippetRepo := new(mocks.ContentSnippetRepositoryMock)

	userID := uuid.New()
	projectID := uuid.New()

	project := &domain.Project{ID: projectID, UserID: userID, Name: "DataSchool", Niche: "Education", SchemaVersion: 3}
	projectRepo.On("GetByID", mock.Anything, projectID.String()).Return(project, nil).Once()

	brandRepo.On("GetByProjectID", mock.Anything, projectID.String()).Return((*domain.BrandProfile)(nil), domain.ErrNotFound).Once()
	productRepo.On("GetByProjectID", mock.Anything, projectID.String()).Return((*domain.ProductProfile)(nil), domain.ErrNotFound).Once()
	snippetRepo.On("ListByProject", mock.Anything, projectID.String()).Return([]*domain.ContentSnippet{}, nil).Once()

	svc := NewCCEService(projectRepo, brandRepo, productRepo, snippetRepo)

	ctxResult, err := svc.BuildPromptContext(context.Background(), userID.String(), projectID.String())
	require.NoError(t, err)
	require.NotNil(t, ctxResult)
	assert.Equal(t, project.Name, ctxResult.Brand.Name)
	assert.Equal(t, "leadgen", ctxResult.Generation.Goal)
	assert.Equal(t, project.SchemaVersion, ctxResult.Project.SchemaVersion)

	projectRepo.AssertExpectations(t)
	brandRepo.AssertExpectations(t)
	productRepo.AssertExpectations(t)
	snippetRepo.AssertExpectations(t)
}

func TestCCEServiceUpsertBrandProfile(t *testing.T) {
	projectRepo := new(mocks.ProjectRepositoryMock)
	brandRepo := new(mocks.BrandProfileRepositoryMock)
	productRepo := new(mocks.ProductProfileRepositoryMock)
	snippetRepo := new(mocks.ContentSnippetRepositoryMock)

	userID := uuid.New()
	projectID := uuid.New()
	project := &domain.Project{ID: projectID, UserID: userID, Name: "Brand"}
	projectRepo.On("GetByID", mock.Anything, projectID.String()).Return(project, nil).Once()

	expected := &domain.BrandProfile{ProjectID: projectID, BrandName: "Brand"}
	brandRepo.On("Upsert", mock.Anything, mock.Anything).Return(expected, nil).Once()

	svc := NewCCEService(projectRepo, brandRepo, productRepo, snippetRepo)
	profile, err := svc.UpsertBrandProfile(context.Background(), userID.String(), projectID.String(), &domain.UpdateBrandProfileRequest{Name: "Brand"})
	require.NoError(t, err)
	assert.NotNil(t, profile)
	assert.Equal(t, "Brand", profile.BrandName)

	projectRepo.AssertExpectations(t)
	brandRepo.AssertExpectations(t)
}

func TestCCEServiceCreateSnippet(t *testing.T) {
	projectRepo := new(mocks.ProjectRepositoryMock)
	brandRepo := new(mocks.BrandProfileRepositoryMock)
	productRepo := new(mocks.ProductProfileRepositoryMock)
	snippetRepo := new(mocks.ContentSnippetRepositoryMock)

	userID := uuid.New()
	projectID := uuid.New()
	project := &domain.Project{ID: projectID, UserID: userID}
	projectRepo.On("GetByID", mock.Anything, projectID.String()).Return(project, nil).Times(2)

	snippet := &domain.ContentSnippet{ProjectID: projectID, Label: "quote"}
	snippetRepo.On("Create", mock.Anything, mock.Anything).Return(snippet, nil).Once()

	snippetRepo.On("ListByProject", mock.Anything, projectID.String()).Return([]*domain.ContentSnippet{snippet}, nil).Once()

	svc := NewCCEService(projectRepo, brandRepo, productRepo, snippetRepo)

	req := &domain.CreateSnippetRequest{Label: "quote", Content: "Keep shipping"}
	created, err := svc.CreateSnippet(context.Background(), userID.String(), projectID.String(), req)
	require.NoError(t, err)
	assert.Equal(t, snippet.Label, created.Label)

	snippets, err := svc.ListSnippets(context.Background(), userID.String(), projectID.String())
	require.NoError(t, err)
	require.Len(t, snippets, 1)

	projectRepo.AssertExpectations(t)
	snippetRepo.AssertExpectations(t)
}
