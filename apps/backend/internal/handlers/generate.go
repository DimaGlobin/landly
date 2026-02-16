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
	"github.com/landly/backend/internal/services"
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
	// CDN support
	HasCDN() bool
	GetCDNURL(remotePath string) string
}

type GenerateHandler struct {
	generateService GenerateService
	publishService  PublishService
	publicBaseProvider services.PublicBaseProvider
}

func NewGenerateHandler(generateService GenerateService, publishService PublishService, publicBaseProvider services.PublicBaseProvider) *GenerateHandler {
	return &GenerateHandler{
		generateService: generateService,
		publishService:  publishService,
		publicBaseProvider: publicBaseProvider,
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
		RespondError(c, domain.ErrUnauthorized)
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		RespondError(c, domain.ErrValidationError.WithMessage("Invalid project ID"))
		return
	}

	var req dto.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err.Error())
		return
	}

	session, err := h.generateService.GenerateSite(c.Request.Context(), userID.String(), projectID.String(), &domain.GenerateRequest{
		Prompt:     req.Prompt,
		PaymentURL: req.PaymentURL,
	})
		if err != nil {
			if domainErr, ok := err.(*domain.Error); ok {
				RespondError(c, domainErr)
				return
			}
			RespondInternalError(c, err)
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
		RespondError(c, domain.ErrUnauthorized)
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		RespondError(c, domain.ErrValidationError.WithMessage("Invalid project ID"))
		return
	}

	preview, err := h.generateService.GetPreview(c.Request.Context(), userID, projectID)
	if err != nil {
		RespondError(c, domain.ErrProjectNotFound)
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
		RespondError(c, domain.ErrUnauthorized)
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		RespondError(c, domain.ErrValidationError.WithMessage("Invalid project ID"))
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
		RespondError(c, domain.ErrUnauthorized)
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		RespondError(c, domain.ErrValidationError.WithMessage("Invalid project ID"))
		return
	}

	var req dto.ChatMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err.Error())
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
		RespondError(c, domain.ErrUnauthorized)
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		RespondError(c, domain.ErrValidationError.WithMessage("Invalid project ID"))
		return
	}

	result, err := h.publishService.PublishSite(c.Request.Context(), userID.String(), projectID.String(), &domain.PublishRequest{
		Domain: "",
		Path:   "",
	})
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			RespondError(c, domainErr)
			return
		}
		// Log unexpected errors before returning 500
		logger.WithContext(c.Request.Context()).Error("publish failed with unexpected error",
			zap.String("project_id", projectID.String()),
			zap.String("user_id", userID.String()),
			zap.Error(err),
		)
		RespondInternalError(c, err)
		return
	}

	if result == nil {
		logger.LogInternalError(c.Request.Context(), "publish returned nil result",
			fmt.Errorf("publish returned nil result"),
			zap.String("project_id", projectID.String()),
			zap.String("user_id", userID.String()),
		)
		RespondInternalErrorWithoutLogging(c)
		return
	}

	if result.Subdomain == "" {
		logger.LogInternalError(c.Request.Context(), "publish returned empty subdomain",
			fmt.Errorf("publish returned empty subdomain"),
			zap.String("project_id", projectID.String()),
			zap.String("user_id", userID.String()),
		)
		RespondInternalErrorWithoutLogging(c)
		return
	}

	var publishedAt string
	if result.LastPublishedAt != nil {
		publishedAt = result.LastPublishedAt.Format("2006-01-02T15:04:05Z")
	}

	c.JSON(http.StatusOK, dto.PublishResponse{
		Subdomain:   result.Subdomain,
		Status:      string(result.Status),
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
		RespondError(c, domain.ErrUnauthorized)
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		RespondError(c, domain.ErrValidationError.WithMessage("Invalid project ID"))
		return
	}

	if err := h.publishService.UnpublishProject(c.Request.Context(), userID, projectID); err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			RespondError(c, domainErr)
			return
		}
		RespondInternalError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ServePublished обрабатывает запросы на опубликованный лендинг
// CDN-first: если CDN настроен, делает редирект на CDN, иначе отдает файл через backend
func (h *GenerateHandler) ServePublished(c *gin.Context) {
	log := logger.WithContext(c.Request.Context())
	
	slug := c.Param("slug")
	asset := strings.TrimPrefix(c.Param("path"), "/")
	
	log.Info("ServePublished request",
		zap.String("slug", slug),
		zap.String("asset", asset),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("query", c.Request.URL.RawQuery))

	if slug == "" {
		log.Warn("empty slug in request")
		c.Status(http.StatusNotFound)
		return
	}

	// Проверяем наличие publishService
	if h.publishService == nil {
		log.Error("publishService is nil")
		c.String(http.StatusInternalServerError, "Internal server error: publish service not initialized")
		return
	}

	// CDN-first serving: если CDN настроен, делаем редирект
	hasCDN := h.publishService.HasCDN()
	log.Debug("CDN check",
		zap.Bool("has_cdn", hasCDN))
	
	if hasCDN {
		// Формируем путь для CDN
		basePath := fmt.Sprintf("sites/%s", slug)
		if asset != "" {
			basePath = fmt.Sprintf("sites/%s/%s", slug, asset)
		}
		cdnURL := h.publishService.GetCDNURL(basePath)
		log.Debug("CDN URL generation",
			zap.String("base_path", basePath),
			zap.String("cdn_url", cdnURL))
		
		if cdnURL != "" {
			log.Info("redirecting to CDN",
				zap.String("cdn_url", cdnURL))
			c.Redirect(http.StatusMovedPermanently, cdnURL)
			return
		}
		log.Warn("CDN enabled but URL is empty, falling back to backend")
	}

	// Fallback: отдаем файл через backend
	log.Debug("serving from backend",
		zap.String("slug", slug),
		zap.String("asset", asset))
	
	reader, contentType, err := h.publishService.ServePublished(c.Request.Context(), slug, asset)
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			// Логируем доменную ошибку с контекстом
			logger.LogDomainError(c.Request.Context(), "failed to serve published asset",
				domainErr,
				zap.String("slug", slug),
				zap.String("asset", asset),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
			)
			c.String(domainErr.HTTPStatus(), domainErr.Message)
			return
		}
		
		// Логируем внутреннюю ошибку с максимальным контекстом
		logger.LogInternalError(c.Request.Context(), "failed to serve published asset",
			err,
			zap.String("slug", slug),
			zap.String("asset", asset),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)
		c.String(http.StatusInternalServerError, "Internal server error")
		return
	}
	
	// Проверяем, что reader не nil перед использованием
	if reader == nil {
		log.Error("ServePublished returned nil reader",
			zap.String("slug", slug),
			zap.String("asset", asset),
			zap.String("content_type", contentType))
		c.String(http.StatusNotFound, "Not found")
		return
	}
	
	log.Debug("file found, serving",
		zap.String("slug", slug),
		zap.String("asset", asset),
		zap.String("content_type", contentType))
	
	defer func() {
		if closeErr := reader.Close(); closeErr != nil {
			log.Warn("failed to close reader", zap.Error(closeErr))
		}
	}()

	if contentType != "" {
		c.Header("Content-Type", contentType)
	}
	// Cache-Control: index.html — short cache; static assets — 5 min
	if asset == "" || asset == "index.html" {
		c.Header("Cache-Control", "public, max-age=60")
	} else {
		c.Header("Cache-Control", "public, max-age=300")
	}

	bytesWritten, err := io.Copy(c.Writer, reader)
	if err != nil {
		log.Error("failed to write published asset",
			zap.String("slug", slug),
			zap.String("asset", asset),
			zap.Int64("bytes_written", bytesWritten),
			zap.Error(err))
		return
	}
	
	log.Info("file served successfully",
		zap.String("slug", slug),
		zap.String("asset", asset),
		zap.Int64("bytes_written", bytesWritten),
		zap.String("content_type", contentType))
}

func respondWithDomainError(c *gin.Context, err error) bool {
	if err == nil {
		return false
	}

	if domainErr, ok := err.(*domain.Error); ok {
		RespondError(c, domainErr)
	} else {
		RespondInternalError(c, err)
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
