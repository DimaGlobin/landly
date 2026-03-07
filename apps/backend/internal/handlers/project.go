package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/landly/backend/internal/handlers/dto"
	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/services"
)

// ProjectService интерфейс для сервиса проектов
type ProjectService interface {
	CreateProject(ctx context.Context, userID string, req *domain.CreateProjectRequest) (*domain.Project, error)
	GetProject(ctx context.Context, userID, projectID string) (*domain.Project, error)
	UpdateProject(ctx context.Context, userID, projectID string, req *domain.UpdateProjectRequest) (*domain.Project, error)
	DeleteProject(ctx context.Context, userID, projectID string) error
	ListProjects(ctx context.Context, userID string, limit, offset int) ([]*domain.Project, error)
}

type ProjectHandler struct {
	projectService    ProjectService
	publishRepo       domain.PublishTargetRepository
	publicBaseProvider services.PublicBaseProvider
}

func NewProjectHandler(projectService ProjectService, publishRepo domain.PublishTargetRepository, publicBaseProvider services.PublicBaseProvider) *ProjectHandler {
	return &ProjectHandler{
		projectService:    projectService,
		publishRepo:       publishRepo,
		publicBaseProvider: publicBaseProvider,
	}
}

func (h *ProjectHandler) getPublishInfo(ctx context.Context, projectID uuid.UUID) *dto.ProjectPublishInfo {
	if h.publishRepo == nil {
		return nil
	}

	target, err := h.publishRepo.GetByProjectID(ctx, projectID.String())
	if err != nil {
		return nil
	}

	if target.Status != domain.PublishStatusPublished {
		return nil
	}

	return &dto.ProjectPublishInfo{
		Status:          target.Status,
		Subdomain:       target.Subdomain,
		LastPublishedAt: target.LastPublishedAt,
	}
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
		RespondError(c, domain.ErrUnauthorized)
		return
	}

	var req dto.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err.Error())
		return
	}

	project, err := h.projectService.CreateProject(c.Request.Context(), userID.String(), &domain.CreateProjectRequest{
		Name:  req.Name,
		Niche: req.Niche,
	})
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			RespondError(c, domainErr)
			return
		}
		RespondInternalError(c, err)
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
		RespondError(c, domain.ErrUnauthorized)
		return
	}

	const defaultLimit = 50
	limit := defaultLimit
	offset := 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	ctx := c.Request.Context()
	projects, err := h.projectService.ListProjects(ctx, userID.String(), limit, offset)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			RespondError(c, domainErr)
			return
		}
		RespondInternalError(c, err)
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
			Publish:   h.getPublishInfo(ctx, p.ID),
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
		RespondError(c, domain.ErrUnauthorized)
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		RespondError(c, domain.ErrValidationError.WithMessage("Invalid project ID"))
		return
	}

	ctx := c.Request.Context()
	project, err := h.projectService.GetProject(ctx, userID.String(), projectID.String())
	if err != nil {
		RespondError(c, domain.ErrProjectNotFound)
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
		Publish:   h.getPublishInfo(ctx, project.ID),
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
		RespondError(c, domain.ErrUnauthorized)
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		RespondError(c, domain.ErrValidationError.WithMessage("Invalid project ID"))
		return
	}

	if err := h.projectService.DeleteProject(c.Request.Context(), userID.String(), projectID.String()); err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			RespondError(c, domainErr)
			return
		}
		RespondInternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
