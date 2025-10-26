package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/landly/backend/internal/handlers/dto"
	domain "github.com/landly/backend/internal/models"
)

// GenerateService интерфейс для сервиса генерации
type GenerateService interface {
	GenerateSite(ctx context.Context, userID, projectID string, req *domain.GenerateRequest) (*domain.GenerationSession, error)
	GetGenerationStatus(ctx context.Context, userID, sessionID string) (*domain.GenerationSession, error)
	GetGenerationResult(ctx context.Context, userID, sessionID string) (*domain.GenerationResult, error)
	GetPreview(ctx context.Context, userID, projectID uuid.UUID) (map[string]interface{}, error)
}

// PublishService интерфейс для сервиса публикации
type PublishService interface {
	PublishSite(ctx context.Context, userID, projectID string, req *domain.PublishRequest) (*domain.PublishTarget, error)
	GetPublishStatus(ctx context.Context, userID, targetID string) (*domain.PublishTarget, error)
	GetPublishedURL(ctx context.Context, userID, targetID string) (string, error)
	UnpublishProject(ctx context.Context, userID, projectID uuid.UUID) error
}

type GenerateHandler struct {
	generateService GenerateService
	publishService  PublishService
	publicBaseURL   string
}

func NewGenerateHandler(generateService GenerateService, publishService PublishService, publicBaseURL string) *GenerateHandler {
	return &GenerateHandler{
		generateService: generateService,
		publishService:  publishService,
		publicBaseURL:   strings.TrimRight(publicBaseURL, "/"),
	}
}

// Generate godoc
// @Summary Generate landing page
// @Tags generate
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body dto.GenerateRequest true "Generate request"
// @Success 200 {object} dto.ProjectResponse
// @Router /v1/projects/{id}/generate [post]
// @Security BearerAuth
func (h *GenerateHandler) Generate(c *gin.Context) {
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

	var req dto.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.generateService.GenerateSite(c.Request.Context(), userID.String(), projectID.String(), &domain.GenerateRequest{
		Prompt:     req.Prompt,
		PaymentURL: req.PaymentURL,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.GenerationSessionResponse{
		ID:        session.ID,
		ProjectID: session.ProjectID,
		Status:    string(session.Status),
		CreatedAt: session.CreatedAt,
	})
}

// GetPreview godoc
// @Summary Get landing page preview
// @Tags generate
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.PreviewResponse
// @Router /v1/projects/{id}/preview [get]
// @Security BearerAuth
func (h *GenerateHandler) GetPreview(c *gin.Context) {
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

	preview, err := h.generateService.GetPreview(c.Request.Context(), userID, projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.PreviewResponse{Schema: preview})
}

// Publish godoc
// @Summary Publish landing page
// @Tags generate
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.PublishResponse
// @Router /v1/projects/{id}/publish [post]
// @Security BearerAuth
func (h *GenerateHandler) Publish(c *gin.Context) {
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

	result, err := h.publishService.PublishSite(c.Request.Context(), userID.String(), projectID.String(), &domain.PublishRequest{
		Domain: "",
		Path:   "",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var publishedAt string
	if result.LastPublishedAt != nil {
		publishedAt = result.LastPublishedAt.Format("2006-01-02T15:04:05Z")
	}

	baseURL := h.publicBaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	publicURL := fmt.Sprintf("%s/%s", baseURL, result.Subdomain)

	c.JSON(http.StatusOK, dto.PublishResponse{
		Subdomain:   result.Subdomain,
		PublicURL:   publicURL,
		PublishedAt: publishedAt,
	})
}

// Unpublish godoc
// @Summary Unpublish landing page
// @Tags generate
// @Param id path string true "Project ID"
// @Success 204
// @Router /v1/projects/{id}/publish [delete]
// @Security BearerAuth
func (h *GenerateHandler) Unpublish(c *gin.Context) {
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

	if err := h.publishService.UnpublishProject(c.Request.Context(), userID, projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
