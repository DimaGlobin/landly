package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/landly/backend/internal/handlers/dto"
	domain "github.com/landly/backend/internal/models"
)

// SchemaVersionService интерфейс для сервиса версий схемы
type SchemaVersionService interface {
	ListVersions(ctx context.Context, userID, projectID string, limit int) ([]*domain.SchemaVersion, error)
	RevertToVersion(ctx context.Context, userID, projectID, versionID string) (*domain.Project, error)
}

type SchemaHandler struct {
	schemaVersionService SchemaVersionService
}

func NewSchemaHandler(schemaVersionService SchemaVersionService) *SchemaHandler {
	return &SchemaHandler{
		schemaVersionService: schemaVersionService,
	}
}

// GetSchemaVersions получает список версий схемы проекта
// @Summary Get schema versions
// @Tags schema
// @Produce json
// @Param id path string true "Project ID"
// @Param limit query int false "Limit (default: 20, max: 100)"
// @Success 200 {object} dto.SchemaVersionsListResponse
// @Router /v1/projects/{id}/schema/versions [get]
// @Security BearerAuth
func (h *SchemaHandler) GetSchemaVersions(c *gin.Context) {
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

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	versions, err := h.schemaVersionService.ListVersions(c.Request.Context(), userID.String(), projectID.String(), limit)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domainErr.HTTPStatus(), gin.H{"error": domainErr.Message})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := make([]dto.SchemaVersionResponse, len(versions))
	for i, v := range versions {
		response[i] = dto.SchemaVersionResponse{
			ID:         v.ID,
			ProjectID:  v.ProjectID,
			CreatedAt:  v.CreatedAt,
			CreatedBy:  v.CreatedBy,
			Source:     string(v.Source),
			TokensUsed: v.TokensUsed,
		}
	}

	c.JSON(http.StatusOK, dto.SchemaVersionsListResponse{
		Versions: response,
		Total:    len(response),
	})
}

// RevertSchema восстанавливает схему проекта из версии
// @Summary Revert schema to version
// @Tags schema
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body dto.RevertSchemaRequest true "Revert request"
// @Success 200 {object} dto.ProjectResponse
// @Router /v1/projects/{id}/schema/revert [post]
// @Security BearerAuth
func (h *SchemaHandler) RevertSchema(c *gin.Context) {
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

	var req dto.RevertSchemaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	project, err := h.schemaVersionService.RevertToVersion(c.Request.Context(), userID.String(), projectID.String(), req.VersionID)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.JSON(domainErr.HTTPStatus(), gin.H{"error": domainErr.Message})
			return
		}
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

