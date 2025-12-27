package services

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/storage/s3"
)

// Mock for ProjectRepository
type mockProjectRepository struct {
	mock.Mock
}

func (m *mockProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *mockProjectRepository) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *mockProjectRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Project, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Project), args.Error(1)
}

func (m *mockProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *mockProjectRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Mock for PublishTargetRepository
type mockPublishTargetRepository struct {
	mock.Mock
}

func (m *mockPublishTargetRepository) GetByProjectID(ctx context.Context, projectID string) (*domain.PublishTarget, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PublishTarget), args.Error(1)
}

// Mock for Publisher
type mockPublisher struct {
	mock.Mock
}

func (m *mockPublisher) Upload(ctx context.Context, localPath, remotePath string) error {
	args := m.Called(ctx, localPath, remotePath)
	return args.Error(0)
}

func (m *mockPublisher) GetPublicURL(remotePath string) string {
	args := m.Called(remotePath)
	return args.String(0)
}

func (m *mockPublisher) GetObject(ctx context.Context, remotePath string) (io.ReadCloser, string, error) {
	args := m.Called(ctx, remotePath)
	if args.Get(0) == nil {
		return nil, args.String(1), args.Error(2)
	}
	return args.Get(0).(io.ReadCloser), args.String(1), args.Error(2)
}

func (m *mockPublisher) DeletePrefix(ctx context.Context, prefix string) error {
	args := m.Called(ctx, prefix)
	return args.Error(0)
}

func (m *mockPublisher) UploadToRelease(ctx context.Context, localPath, basePath string, releaseID uuid.UUID) error {
	args := m.Called(ctx, localPath, basePath, releaseID)
	return args.Error(0)
}

func (m *mockPublisher) CreateManifest(ctx context.Context, localPath string) (*s3.ReleaseManifest, error) {
	args := m.Called(ctx, localPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.ReleaseManifest), args.Error(1)
}

func (m *mockPublisher) UploadIncremental(ctx context.Context, localPath, basePath string, releaseID uuid.UUID, previousManifest *s3.ReleaseManifest) (*s3.ReleaseManifest, error) {
	args := m.Called(ctx, localPath, basePath, releaseID, previousManifest)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*s3.ReleaseManifest), args.Error(1)
}

func (m *mockPublisher) SetCurrentRelease(ctx context.Context, basePath string, releaseID uuid.UUID) error {
	args := m.Called(ctx, basePath, releaseID)
	return args.Error(0)
}

func (m *mockPublisher) GetCurrentReleasePath(basePath string) string {
	args := m.Called(basePath)
	return args.String(0)
}

func (m *mockPublisher) GetReleasePath(basePath string, releaseID uuid.UUID) string {
	args := m.Called(basePath, releaseID)
	return args.String(0)
}

func (m *mockPublisher) HasCDN() bool {
	args := m.Called()
	return args.Bool(0)
}

func TestProjectService_CreateProject_Success(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProjectRepository)
	publishTargetRepo := new(mockPublishTargetRepository)
	publisher := new(mockPublisher)
	service := NewProjectService(repo, publishTargetRepo, publisher)

	userID := uuid.New()

	repo.On("Create", ctx, mock.MatchedBy(func(p *domain.Project) bool {
		return p.Name == "My Project" && p.Niche == "SaaS" && p.UserID == userID
	})).Return(nil)

	project, err := service.CreateProject(ctx, userID.String(), &domain.CreateProjectRequest{
		Name:  "My Project",
		Niche: "SaaS",
	})

	require.NoError(t, err)
	require.NotNil(t, project)
	assert.Equal(t, "My Project", project.Name)
	assert.Equal(t, "SaaS", project.Niche)
	assert.Equal(t, userID, project.UserID)
	assert.Equal(t, domain.ProjectStatusDraft, project.Status)

	repo.AssertExpectations(t)
}

func TestProjectService_CreateProject_InvalidUserID(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProjectRepository)
	publishTargetRepo := new(mockPublishTargetRepository)
	publisher := new(mockPublisher)
	service := NewProjectService(repo, publishTargetRepo, publisher)

	project, err := service.CreateProject(ctx, "invalid-uuid", &domain.CreateProjectRequest{
		Name:  "My Project",
		Niche: "SaaS",
	})

	assert.Nil(t, project)
	require.Error(t, err)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, "BAD_REQUEST", domainErr.Code)
}

func TestProjectService_CreateProject_MissingName(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProjectRepository)
	publishTargetRepo := new(mockPublishTargetRepository)
	publisher := new(mockPublisher)
	service := NewProjectService(repo, publishTargetRepo, publisher)

	userID := uuid.New()

	project, err := service.CreateProject(ctx, userID.String(), &domain.CreateProjectRequest{
		Name:  "",
		Niche: "SaaS",
	})

	assert.Nil(t, project)
	require.Error(t, err)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, "BAD_REQUEST", domainErr.Code)
	assert.Contains(t, domainErr.Message, "name")
}

func TestProjectService_CreateProject_MissingNiche(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProjectRepository)
	publishTargetRepo := new(mockPublishTargetRepository)
	publisher := new(mockPublisher)
	service := NewProjectService(repo, publishTargetRepo, publisher)

	userID := uuid.New()

	project, err := service.CreateProject(ctx, userID.String(), &domain.CreateProjectRequest{
		Name:  "My Project",
		Niche: "",
	})

	assert.Nil(t, project)
	require.Error(t, err)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, "BAD_REQUEST", domainErr.Code)
	assert.Contains(t, domainErr.Message, "niche")
}

func TestProjectService_CreateProject_RepositoryError(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProjectRepository)
	publishTargetRepo := new(mockPublishTargetRepository)
	publisher := new(mockPublisher)
	service := NewProjectService(repo, publishTargetRepo, publisher)

	userID := uuid.New()

	repo.On("Create", ctx, mock.Anything).Return(errors.New("db error"))

	project, err := service.CreateProject(ctx, userID.String(), &domain.CreateProjectRequest{
		Name:  "My Project",
		Niche: "SaaS",
	})

	assert.Nil(t, project)
	require.Error(t, err)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, "INTERNAL_ERROR", domainErr.Code)

	repo.AssertExpectations(t)
}

func TestProjectService_GetProject_Success(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProjectRepository)
	publishTargetRepo := new(mockPublishTargetRepository)
	publisher := new(mockPublisher)
	service := NewProjectService(repo, publishTargetRepo, publisher)

	userID := uuid.New()
	projectID := uuid.New()
	existingProject := &domain.Project{
		ID:     projectID,
		UserID: userID,
		Name:   "Test Project",
	}

	repo.On("GetByID", ctx, projectID.String()).Return(existingProject, nil)

	project, err := service.GetProject(ctx, userID.String(), projectID.String())

	require.NoError(t, err)
	assert.Equal(t, existingProject, project)

	repo.AssertExpectations(t)
}

func TestProjectService_GetProject_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProjectRepository)
	publishTargetRepo := new(mockPublishTargetRepository)
	publisher := new(mockPublisher)
	service := NewProjectService(repo, publishTargetRepo, publisher)

	userID := uuid.New()
	projectID := uuid.New()

	repo.On("GetByID", ctx, projectID.String()).Return(nil, errors.New("not found"))

	project, err := service.GetProject(ctx, userID.String(), projectID.String())

	assert.Nil(t, project)
	require.Error(t, err)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, "NOT_FOUND", domainErr.Code)

	repo.AssertExpectations(t)
}

func TestProjectService_GetProject_Forbidden(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProjectRepository)
	publishTargetRepo := new(mockPublishTargetRepository)
	publisher := new(mockPublisher)
	service := NewProjectService(repo, publishTargetRepo, publisher)

	ownerID := uuid.New()
	otherUserID := uuid.New()
	projectID := uuid.New()
	existingProject := &domain.Project{
		ID:     projectID,
		UserID: ownerID, // Different user owns the project
		Name:   "Test Project",
	}

	repo.On("GetByID", ctx, projectID.String()).Return(existingProject, nil)

	project, err := service.GetProject(ctx, otherUserID.String(), projectID.String())

	assert.Nil(t, project)
	require.Error(t, err)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, "FORBIDDEN", domainErr.Code)

	repo.AssertExpectations(t)
}

func TestProjectService_ListProjects_Success(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProjectRepository)
	publishTargetRepo := new(mockPublishTargetRepository)
	publisher := new(mockPublisher)
	service := NewProjectService(repo, publishTargetRepo, publisher)

	userID := uuid.New()
	projects := []*domain.Project{
		{ID: uuid.New(), UserID: userID, Name: "Project 1"},
		{ID: uuid.New(), UserID: userID, Name: "Project 2"},
	}

	repo.On("GetByUserID", ctx, userID.String()).Return(projects, nil)

	result, err := service.ListProjects(ctx, userID.String())

	require.NoError(t, err)
	assert.Len(t, result, 2)

	repo.AssertExpectations(t)
}

func TestProjectService_ListProjects_EmptyList(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProjectRepository)
	publishTargetRepo := new(mockPublishTargetRepository)
	publisher := new(mockPublisher)
	service := NewProjectService(repo, publishTargetRepo, publisher)

	userID := uuid.New()

	repo.On("GetByUserID", ctx, userID.String()).Return([]*domain.Project{}, nil)

	result, err := service.ListProjects(ctx, userID.String())

	require.NoError(t, err)
	assert.Empty(t, result)

	repo.AssertExpectations(t)
}

func TestProjectService_ListProjects_RepositoryError(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProjectRepository)
	publishTargetRepo := new(mockPublishTargetRepository)
	publisher := new(mockPublisher)
	service := NewProjectService(repo, publishTargetRepo, publisher)

	userID := uuid.New()

	repo.On("GetByUserID", ctx, userID.String()).Return(nil, errors.New("db error"))

	result, err := service.ListProjects(ctx, userID.String())

	assert.Nil(t, result)
	require.Error(t, err)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, "INTERNAL_ERROR", domainErr.Code)

	repo.AssertExpectations(t)
}

func TestProjectService_UpdateProject_Success(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProjectRepository)
	publishTargetRepo := new(mockPublishTargetRepository)
	publisher := new(mockPublisher)
	service := NewProjectService(repo, publishTargetRepo, publisher)

	userID := uuid.New()
	projectID := uuid.New()
	existingProject := &domain.Project{
		ID:     projectID,
		UserID: userID,
		Name:   "Old Name",
		Niche:  "Old Niche",
	}

	repo.On("GetByID", ctx, projectID.String()).Return(existingProject, nil)
	repo.On("Update", ctx, mock.MatchedBy(func(p *domain.Project) bool {
		return p.Name == "New Name" && p.Niche == "New Niche"
	})).Return(nil)

	project, err := service.UpdateProject(ctx, userID.String(), projectID.String(), &domain.UpdateProjectRequest{
		Name:  "New Name",
		Niche: "New Niche",
	})

	require.NoError(t, err)
	assert.Equal(t, "New Name", project.Name)
	assert.Equal(t, "New Niche", project.Niche)

	repo.AssertExpectations(t)
}

func TestProjectService_UpdateProject_PartialUpdate(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProjectRepository)
	publishTargetRepo := new(mockPublishTargetRepository)
	publisher := new(mockPublisher)
	service := NewProjectService(repo, publishTargetRepo, publisher)

	userID := uuid.New()
	projectID := uuid.New()
	existingProject := &domain.Project{
		ID:     projectID,
		UserID: userID,
		Name:   "Old Name",
		Niche:  "Old Niche",
	}

	repo.On("GetByID", ctx, projectID.String()).Return(existingProject, nil)
	repo.On("Update", ctx, mock.MatchedBy(func(p *domain.Project) bool {
		return p.Name == "New Name" && p.Niche == "Old Niche" // Niche unchanged
	})).Return(nil)

	project, err := service.UpdateProject(ctx, userID.String(), projectID.String(), &domain.UpdateProjectRequest{
		Name: "New Name",
		// Niche is empty - should not update
	})

	require.NoError(t, err)
	assert.Equal(t, "New Name", project.Name)
	assert.Equal(t, "Old Niche", project.Niche)

	repo.AssertExpectations(t)
}

func TestProjectService_UpdateProject_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProjectRepository)
	publishTargetRepo := new(mockPublishTargetRepository)
	publisher := new(mockPublisher)
	service := NewProjectService(repo, publishTargetRepo, publisher)

	userID := uuid.New()
	projectID := uuid.New()

	repo.On("GetByID", ctx, projectID.String()).Return(nil, errors.New("not found"))

	project, err := service.UpdateProject(ctx, userID.String(), projectID.String(), &domain.UpdateProjectRequest{
		Name: "New Name",
	})

	assert.Nil(t, project)
	require.Error(t, err)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, "NOT_FOUND", domainErr.Code)

	repo.AssertExpectations(t)
}

func TestProjectService_UpdateProject_Forbidden(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProjectRepository)
	publishTargetRepo := new(mockPublishTargetRepository)
	publisher := new(mockPublisher)
	service := NewProjectService(repo, publishTargetRepo, publisher)

	ownerID := uuid.New()
	otherUserID := uuid.New()
	projectID := uuid.New()
	existingProject := &domain.Project{
		ID:     projectID,
		UserID: ownerID, // Different user owns the project
		Name:   "Test Project",
	}

	repo.On("GetByID", ctx, projectID.String()).Return(existingProject, nil)

	project, err := service.UpdateProject(ctx, otherUserID.String(), projectID.String(), &domain.UpdateProjectRequest{
		Name: "New Name",
	})

	assert.Nil(t, project)
	require.Error(t, err)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, "FORBIDDEN", domainErr.Code)

	repo.AssertExpectations(t)
}

func TestProjectService_DeleteProject_Success(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProjectRepository)
	publishTargetRepo := new(mockPublishTargetRepository)
	publisher := new(mockPublisher)
	service := NewProjectService(repo, publishTargetRepo, publisher)

	userID := uuid.New()
	projectID := uuid.New()
	existingProject := &domain.Project{
		ID:     projectID,
		UserID: userID,
		Name:   "Test Project",
	}

	target := &domain.PublishTarget{
		ID:        uuid.New(),
		ProjectID: projectID,
		Subdomain: "test-project-12345678",
	}

	repo.On("GetByID", ctx, projectID.String()).Return(existingProject, nil)
	publishTargetRepo.On("GetByProjectID", ctx, projectID.String()).Return(target, nil)
	publisher.On("DeletePrefix", ctx, "sites/test-project-12345678").Return(nil)
	repo.On("Delete", ctx, projectID.String()).Return(nil)

	err := service.DeleteProject(ctx, userID.String(), projectID.String())

	require.NoError(t, err)
	repo.AssertExpectations(t)
	publishTargetRepo.AssertExpectations(t)
	publisher.AssertExpectations(t)
}

func TestProjectService_DeleteProject_NotFound(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProjectRepository)
	publishTargetRepo := new(mockPublishTargetRepository)
	publisher := new(mockPublisher)
	service := NewProjectService(repo, publishTargetRepo, publisher)

	userID := uuid.New()
	projectID := uuid.New()

	repo.On("GetByID", ctx, projectID.String()).Return(nil, errors.New("not found"))

	err := service.DeleteProject(ctx, userID.String(), projectID.String())

	require.Error(t, err)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, "NOT_FOUND", domainErr.Code)

	repo.AssertExpectations(t)
}

func TestProjectService_DeleteProject_Forbidden(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProjectRepository)
	publishTargetRepo := new(mockPublishTargetRepository)
	publisher := new(mockPublisher)
	service := NewProjectService(repo, publishTargetRepo, publisher)

	ownerID := uuid.New()
	otherUserID := uuid.New()
	projectID := uuid.New()
	existingProject := &domain.Project{
		ID:     projectID,
		UserID: ownerID,
		Name:   "Test Project",
	}

	repo.On("GetByID", ctx, projectID.String()).Return(existingProject, nil)

	err := service.DeleteProject(ctx, otherUserID.String(), projectID.String())

	require.Error(t, err)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, "FORBIDDEN", domainErr.Code)

	repo.AssertExpectations(t)
}

func TestProjectService_DeleteProject_RepositoryError(t *testing.T) {
	ctx := context.Background()
	repo := new(mockProjectRepository)
	publishTargetRepo := new(mockPublishTargetRepository)
	publisher := new(mockPublisher)
	service := NewProjectService(repo, publishTargetRepo, publisher)

	userID := uuid.New()
	projectID := uuid.New()
	existingProject := &domain.Project{
		ID:     projectID,
		UserID: userID,
		Name:   "Test Project",
	}

	repo.On("GetByID", ctx, projectID.String()).Return(existingProject, nil)
	// GetByProjectID вызывается перед удалением для проверки наличия publish target
	publishTargetRepo.On("GetByProjectID", ctx, projectID.String()).Return(nil, errors.New("not found"))
	repo.On("Delete", ctx, projectID.String()).Return(errors.New("db error"))

	err := service.DeleteProject(ctx, userID.String(), projectID.String())

	require.Error(t, err)

	domainErr, ok := err.(*domain.Error)
	require.True(t, ok)
	assert.Equal(t, "INTERNAL_ERROR", domainErr.Code)

	repo.AssertExpectations(t)
	publishTargetRepo.AssertExpectations(t)
}
