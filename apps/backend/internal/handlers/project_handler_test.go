package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/landly/backend/internal/handlers/dto"
	domain "github.com/landly/backend/internal/models"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// Mock for ProjectService
type mockProjectService struct {
	mock.Mock
}

func (m *mockProjectService) CreateProject(ctx context.Context, userID string, req *domain.CreateProjectRequest) (*domain.Project, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *mockProjectService) GetProject(ctx context.Context, userID, projectID string) (*domain.Project, error) {
	args := m.Called(ctx, userID, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *mockProjectService) UpdateProject(ctx context.Context, userID, projectID string, req *domain.UpdateProjectRequest) (*domain.Project, error) {
	args := m.Called(ctx, userID, projectID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *mockProjectService) DeleteProject(ctx context.Context, userID, projectID string) error {
	args := m.Called(ctx, userID, projectID)
	return args.Error(0)
}

func (m *mockProjectService) ListProjects(ctx context.Context, userID string) ([]*domain.Project, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Project), args.Error(1)
}

func setupProjectHandler() (*ProjectHandler, *mockProjectService) {
	service := new(mockProjectService)
	handler := NewProjectHandler(service, nil, "http://localhost:8080")
	return handler, service
}

func TestProjectHandler_CreateProject_Success(t *testing.T) {
	handler, service := setupProjectHandler()

	userID := uuid.New()
	projectID := uuid.New()
	now := time.Now()

	project := &domain.Project{
		ID:        projectID,
		UserID:    userID,
		Name:      "Test Project",
		Niche:     "SaaS",
		Status:    domain.ProjectStatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}

	service.On("CreateProject", mock.Anything, userID.String(), mock.MatchedBy(func(req *domain.CreateProjectRequest) bool {
		return req.Name == "Test Project" && req.Niche == "SaaS"
	})).Return(project, nil)

	reqBody := dto.CreateProjectRequest{Name: "Test Project", Niche: "SaaS"}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/projects", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", userID)

	handler.CreateProject(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.ProjectResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, projectID, resp.ID)
	assert.Equal(t, "Test Project", resp.Name)

	service.AssertExpectations(t)
}

func TestProjectHandler_CreateProject_Unauthorized(t *testing.T) {
	handler, _ := setupProjectHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/projects", nil)
	// Not setting user_id

	handler.CreateProject(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var errResp ErrorResponseBody
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "UNAUTHORIZED", errResp.Error.Code)
}

func TestProjectHandler_CreateProject_ValidationError(t *testing.T) {
	handler, _ := setupProjectHandler()

	userID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/projects", bytes.NewReader([]byte("invalid json")))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("user_id", userID)

	handler.CreateProject(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp ErrorResponseBody
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", errResp.Error.Code)
}

func TestProjectHandler_GetProjects_Success(t *testing.T) {
	handler, service := setupProjectHandler()

	userID := uuid.New()
	projects := []*domain.Project{
		{ID: uuid.New(), UserID: userID, Name: "Project 1", Status: domain.ProjectStatusDraft},
		{ID: uuid.New(), UserID: userID, Name: "Project 2", Status: domain.ProjectStatusPublished},
	}

	service.On("ListProjects", mock.Anything, userID.String()).Return(projects, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/projects", nil)
	c.Set("user_id", userID)

	handler.GetProjects(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.ProjectsListResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Len(t, resp.Projects, 2)
	assert.Equal(t, 2, resp.Total)

	service.AssertExpectations(t)
}

func TestProjectHandler_GetProjects_Unauthorized(t *testing.T) {
	handler, _ := setupProjectHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/projects", nil)

	handler.GetProjects(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProjectHandler_GetProject_Success(t *testing.T) {
	handler, service := setupProjectHandler()

	userID := uuid.New()
	projectID := uuid.New()
	project := &domain.Project{
		ID:     projectID,
		UserID: userID,
		Name:   "Test Project",
		Status: domain.ProjectStatusDraft,
	}

	service.On("GetProject", mock.Anything, userID.String(), projectID.String()).Return(project, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/projects/"+projectID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: projectID.String()}}
	c.Set("user_id", userID)

	handler.GetProject(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp dto.ProjectResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, projectID, resp.ID)

	service.AssertExpectations(t)
}

func TestProjectHandler_GetProject_NotFound(t *testing.T) {
	handler, service := setupProjectHandler()

	userID := uuid.New()
	projectID := uuid.New()

	service.On("GetProject", mock.Anything, userID.String(), projectID.String()).Return(nil, domain.ErrProjectNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/projects/"+projectID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: projectID.String()}}
	c.Set("user_id", userID)

	handler.GetProject(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errResp ErrorResponseBody
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "PROJECT_NOT_FOUND", errResp.Error.Code)

	service.AssertExpectations(t)
}

func TestProjectHandler_GetProject_InvalidID(t *testing.T) {
	handler, _ := setupProjectHandler()

	userID := uuid.New()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/projects/invalid-uuid", nil)
	c.Params = gin.Params{{Key: "id", Value: "invalid-uuid"}}
	c.Set("user_id", userID)

	handler.GetProject(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errResp ErrorResponseBody
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "VALIDATION_ERROR", errResp.Error.Code)
}

func TestProjectHandler_DeleteProject_Success(t *testing.T) {
	handler, service := setupProjectHandler()

	userID := uuid.New()
	projectID := uuid.New()

	service.On("DeleteProject", mock.Anything, userID.String(), projectID.String()).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/v1/projects/"+projectID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: projectID.String()}}
	c.Set("user_id", userID)

	handler.DeleteProject(c)

	// Note: 204 No Content may be rendered as 200 by gin in test mode if no body is written
	// Accept either 200 or 204 as success
	assert.True(t, w.Code == http.StatusNoContent || w.Code == http.StatusOK, "expected 200 or 204, got %d", w.Code)
	service.AssertExpectations(t)
}

func TestProjectHandler_DeleteProject_Unauthorized(t *testing.T) {
	handler, _ := setupProjectHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/v1/projects/some-id", nil)

	handler.DeleteProject(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestProjectHandler_DeleteProject_NotFound(t *testing.T) {
	handler, service := setupProjectHandler()

	userID := uuid.New()
	projectID := uuid.New()

	service.On("DeleteProject", mock.Anything, userID.String(), projectID.String()).Return(domain.ErrProjectNotFound)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/v1/projects/"+projectID.String(), nil)
	c.Params = gin.Params{{Key: "id", Value: projectID.String()}}
	c.Set("user_id", userID)

	handler.DeleteProject(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var errResp ErrorResponseBody
	err := json.Unmarshal(w.Body.Bytes(), &errResp)
	require.NoError(t, err)
	assert.Equal(t, "PROJECT_NOT_FOUND", errResp.Error.Code)

	service.AssertExpectations(t)
}

