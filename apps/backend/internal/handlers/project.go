package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/landly/backend/internal/handlers/dto"
	domain "github.com/landly/backend/internal/models"
)

// ProjectService интерфейс для сервиса проектов
type ProjectService interface {
	CreateProject(ctx context.Context, userID string, req *domain.CreateProjectRequest) (*domain.Project, error)
	GetProject(ctx context.Context, userID, projectID string) (*domain.Project, error)
	UpdateProject(ctx context.Context, userID, projectID string, req *domain.UpdateProjectRequest) (*domain.Project, error)
	DeleteProject(ctx context.Context, userID, projectID string) error
	ListProjects(ctx context.Context, userID string) ([]*domain.Project, error)
}

type ProjectHandler struct {
	projectService ProjectService
}

func NewProjectHandler(projectService ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

// CreateProject godoc
// @Summary Create new project
// @Tags projects
// @Accept json
// @Produce json
// @Param request body dto.CreateProjectRequest true "Create project request"
// @Success 200 {object} dto.ProjectResponse
// @Router /v1/projects [post]
// @Security BearerAuth
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req dto.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.projectService.CreateProject(c.Request.Context(), userID.String(), &domain.CreateProjectRequest{
		Name:  req.Name,
		Niche: req.Niche,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.ProjectResponse{
		ID:        project.ID,
		UserID:    project.UserID,
		Name:      project.Name,
		Niche:     project.Niche,
		Status:    string(project.Status),
		CreatedAt: project.CreatedAt,
		UpdatedAt: project.UpdatedAt,
	})
}

// GetProjects godoc
// @Summary Get user projects
// @Tags projects
// @Produce json
// @Success 200 {object} dto.ProjectsListResponse
// @Router /v1/projects [get]
// @Security BearerAuth
func (h *ProjectHandler) GetProjects(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projects, err := h.projectService.ListProjects(c.Request.Context(), userID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]dto.ProjectResponse, len(projects))
	for i, p := range projects {
		response[i] = dto.ProjectResponse{
			ID:        p.ID,
			UserID:    p.UserID,
			Name:      p.Name,
			Niche:     p.Niche,
			Status:    string(p.Status),
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, dto.ProjectsListResponse{
		Projects: response,
		Total:    len(response),
	})
}

// GetProject godoc
// @Summary Get project by ID
// @Tags projects
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ProjectResponse
// @Router /v1/projects/{id} [get]
// @Security BearerAuth
func (h *ProjectHandler) GetProject(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	project, err := h.projectService.GetProject(c.Request.Context(), userID.String(), projectID.String())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	c.JSON(http.StatusOK, dto.ProjectResponse{
		ID:        project.ID,
		UserID:    project.UserID,
		Name:      project.Name,
		Niche:     project.Niche,
		Status:    string(project.Status),
		CreatedAt: project.CreatedAt,
		UpdatedAt: project.UpdatedAt,
	})
}

// DeleteProject godoc
// @Summary Delete project
// @Tags projects
// @Param id path string true "Project ID"
// @Success 204
// @Router /v1/projects/{id} [delete]
// @Security BearerAuth
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	if err := h.projectService.DeleteProject(c.Request.Context(), userID.String(), projectID.String()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
