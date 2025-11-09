package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/services"
)

type CCEService interface {
	UpsertBrandProfile(ctx context.Context, userID, projectID string, req *domain.UpdateBrandProfileRequest) (*domain.BrandProfile, error)
	UpsertProductProfile(ctx context.Context, userID, projectID string, req *domain.UpdateProductProfileRequest) (*domain.ProductProfile, error)
	CreateSnippet(ctx context.Context, userID, projectID string, req *domain.CreateSnippetRequest) (*domain.ContentSnippet, error)
	ListSnippets(ctx context.Context, userID, projectID string) ([]*domain.ContentSnippet, error)
	BuildPromptContext(ctx context.Context, userID, projectID string) (*services.PromptContext, error)
}

type CCEHandler struct {
	service CCEService
}

func NewCCEHandler(service CCEService) *CCEHandler {
	return &CCEHandler{service: service}
}

func (h *CCEHandler) UpdateBrandProfile(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projectID := strings.TrimSpace(c.Param("id"))
	var req domain.UpdateBrandProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.service.UpsertBrandProfile(c.Request.Context(), userID.String(), projectID, &req)
	if respondWithDomainError(c, err) {
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *CCEHandler) UpdateProductProfile(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projectID := strings.TrimSpace(c.Param("id"))
	var req domain.UpdateProductProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	profile, err := h.service.UpsertProductProfile(c.Request.Context(), userID.String(), projectID, &req)
	if respondWithDomainError(c, err) {
		return
	}

	c.JSON(http.StatusOK, profile)
}

func (h *CCEHandler) CreateSnippet(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projectID := strings.TrimSpace(c.Param("id"))
	var req domain.CreateSnippetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	snippet, err := h.service.CreateSnippet(c.Request.Context(), userID.String(), projectID, &req)
	if respondWithDomainError(c, err) {
		return
	}

	c.JSON(http.StatusOK, snippet)
}

func (h *CCEHandler) ListSnippets(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projectID := strings.TrimSpace(c.Param("id"))
	snippets, err := h.service.ListSnippets(c.Request.Context(), userID.String(), projectID)
	if respondWithDomainError(c, err) {
		return
	}

	c.JSON(http.StatusOK, gin.H{"snippets": snippets})
}

func (h *CCEHandler) GetContext(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projectID := strings.TrimSpace(c.Param("id"))
	ctx, err := h.service.BuildPromptContext(c.Request.Context(), userID.String(), projectID)
	if respondWithDomainError(c, err) {
		return
	}

	c.JSON(http.StatusOK, ctx)
}
