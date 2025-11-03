package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/landly/backend/internal/handlers/dto"
	"github.com/landly/backend/internal/logger"
	domain "github.com/landly/backend/internal/models"
	"go.uber.org/zap"
)

// GenerateService интерфейс для сервиса генерации
type GenerateService interface {
	GenerateSite(ctx context.Context, userID, projectID string, req *domain.GenerateRequest) (*domain.GenerationSession, error)
	GetGenerationStatus(ctx context.Context, userID, sessionID string) (*domain.GenerationSession, error)
	GetGenerationResult(ctx context.Context, userID, sessionID string) (*domain.GenerationResult, error)
	GetPreview(ctx context.Context, userID, projectID uuid.UUID) (map[string]interface{}, error)
	GetChatHistory(ctx context.Context, userID, projectID string) (*domain.GenerationSession, []*domain.GenerationMessage, error)
	SendChatMessage(ctx context.Context, userID, projectID, content string) (*domain.GenerationSession, []*domain.GenerationMessage, error)
}

// PublishService интерфейс для сервиса публикации
type PublishService interface {
	PublishSite(ctx context.Context, userID, projectID string, req *domain.PublishRequest) (*domain.PublishTarget, error)
	GetPublishStatus(ctx context.Context, userID, targetID string) (*domain.PublishTarget, error)
	GetPublishedURL(ctx context.Context, userID, targetID string) (string, error)
	UnpublishProject(ctx context.Context, userID, projectID uuid.UUID) error
	ServePublished(ctx context.Context, subdomain, assetPath string) (io.ReadCloser, string, error)
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

// GetChat godoc
// @Summary Get chat history for project
// @Tags generate
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.ChatHistoryResponse
// @Router /v1/projects/{id}/chat [get]
// @Security BearerAuth
func (h *GenerateHandler) GetChat(c *gin.Context) {
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

	session, messages, err := h.generateService.GetChatHistory(c.Request.Context(), userID.String(), projectID.String())
	if respondWithDomainError(c, err) {
		return
	}

	c.JSON(http.StatusOK, dto.ChatHistoryResponse{
		Session:  toChatSessionResponse(session),
		Messages: toChatMessagesResponse(messages),
	})
}

// SendChat godoc
// @Summary Send chat message for project generation
// @Tags generate
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body dto.ChatMessageRequest true "Chat message"
// @Success 200 {object} dto.ChatHistoryResponse
// @Router /v1/projects/{id}/chat [post]
// @Security BearerAuth
func (h *GenerateHandler) SendChat(c *gin.Context) {
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

	var req dto.ChatMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, messages, err := h.generateService.SendChatMessage(c.Request.Context(), userID.String(), projectID.String(), req.Content)
	if respondWithDomainError(c, err) {
		return
	}

	c.JSON(http.StatusOK, dto.ChatHistoryResponse{
		Session:  toChatSessionResponse(session),
		Messages: toChatMessagesResponse(messages),
	})
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

	publicURL := fmt.Sprintf("%s/sites/%s", baseURL, result.Subdomain)

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

// ServePublished обрабатывает запросы на опубликованный лендинг
func (h *GenerateHandler) ServePublished(c *gin.Context) {
	slug := c.Param("slug")
	asset := strings.TrimPrefix(c.Param("path"), "/")

	if slug == "" {
		c.Status(http.StatusNotFound)
		return
	}

	reader, contentType, err := h.publishService.ServePublished(c.Request.Context(), slug, asset)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.String(domainErr.HTTPStatus(), domainErr.Message)
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer reader.Close()

	if contentType != "" {
		c.Header("Content-Type", contentType)
	}

	if _, err := io.Copy(c.Writer, reader); err != nil {
		if errHandler := c.Error(err); errHandler != nil {
			logger.WithContext(c.Request.Context()).Error("failed to write published asset", zap.Error(errHandler))
		}
	}
}

func (h *GenerateHandler) ServePublishedLegacy(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" || isReservedSlug(slug) {
		c.Status(http.StatusNotFound)
		return
	}

	reader, contentType, err := h.publishService.ServePublished(c.Request.Context(), slug, "")
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			c.String(domainErr.HTTPStatus(), domainErr.Message)
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	defer reader.Close()

	if contentType != "" {
		c.Header("Content-Type", contentType)
	}

	if _, err := io.Copy(c.Writer, reader); err != nil {
		if errHandler := c.Error(err); errHandler != nil {
			logger.WithContext(c.Request.Context()).Error("failed to write legacy published asset", zap.Error(errHandler))
		}
	}
}

func isReservedSlug(slug string) bool {
	reserved := map[string]struct{}{
		"health":  {},
		"healthz": {},
		"readyz":  {},
		"sites":   {},
		"v1":      {},
	}

	_, ok := reserved[strings.ToLower(slug)]
	return ok
}

func respondWithDomainError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	if domainErr, ok := err.(*domain.Error); ok {
		c.JSON(domainErr.HTTPStatus(), gin.H{"error": domainErr.Message})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	return true
}

func toChatSessionResponse(session *domain.GenerationSession) dto.ChatSessionResponse {
	response := dto.ChatSessionResponse{}
	if session == nil {
		return response
	}

	response.ID = session.ID
	response.ProjectID = session.ProjectID
	response.Status = session.Status
	response.SchemaJSON = session.SchemaJSON
	response.CompletedAt = session.CompletedAt
	response.CreatedAt = session.CreatedAt
	response.UpdatedAt = session.UpdatedAt

	return response
}

func toChatMessagesResponse(messages []*domain.GenerationMessage) []dto.ChatMessageResponse {
	if len(messages) == 0 {
		return []dto.ChatMessageResponse{}
	}

	result := make([]dto.ChatMessageResponse, len(messages))
	for i, msg := range messages {
		if msg == nil {
			continue
		}
		result[i] = dto.ChatMessageResponse{
			ID:         msg.ID,
			Role:       msg.Role,
			Content:    msg.Content,
			Metadata:   msg.Metadata,
			TokensUsed: msg.TokensUsed,
			CreatedAt:  msg.CreatedAt,
		}
	}

	return result
}
