package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/landly/backend/internal/handlers/dto"
	"github.com/landly/backend/internal/logger"
	domain "github.com/landly/backend/internal/models"
	"go.uber.org/zap"
)

// AnalyticsService интерфейс для сервиса аналитики
type AnalyticsService interface {
	GetProjectAnalytics(ctx context.Context, userID, projectID string) (*domain.ProjectAnalytics, error)
	GetSiteAnalytics(ctx context.Context, userID, targetID string) (*domain.SiteAnalytics, error)
	TrackEvent(ctx context.Context, req *domain.TrackEventRequest) error
}

type AnalyticsHandler struct {
	analyticsService AnalyticsService
}

func NewAnalyticsHandler(analyticsService AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: analyticsService}
}

// TrackEvent godoc
// @Summary Track analytics event
// @Tags analytics
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param request body dto.TrackEventRequest true "Track event request"
// @Success 204
// @Router /v1/analytics/{id}/event [post]
func (h *AnalyticsHandler) TrackEvent(c *gin.Context) {
	var req dto.TrackEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, err.Error())
		return
	}

	err := h.analyticsService.TrackEvent(c.Request.Context(), &domain.TrackEventRequest{
		EventType: req.EventType,
		Path:      req.Path,
		Referrer:  req.Referrer,
	})

	if err != nil {
		logger.WithContext(c.Request.Context()).Warn("failed to track analytics event", zap.Error(err))
	}

	c.Status(http.StatusNoContent)
}

// GetStats godoc
// @Summary Get analytics stats
// @Tags analytics
// @Produce json
// @Param id path string true "Project ID"
// @Success 200 {object} dto.AnalyticsStatsResponse
// @Router /v1/analytics/{id}/stats [get]
// @Security BearerAuth
func (h *AnalyticsHandler) GetStats(c *gin.Context) {
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

	stats, err := h.analyticsService.GetProjectAnalytics(c.Request.Context(), userID.String(), projectID.String())
	if err != nil {
		if domainErr, ok := err.(*domain.Error); ok {
			RespondError(c, domainErr)
			return
		}
		RespondInternalError(c)
		return
	}

	c.JSON(http.StatusOK, dto.AnalyticsStatsResponse{
		ProjectID:      projectID,
		TotalPageViews: int64(stats.TotalPageViews),
		TotalCTAClicks: int64(stats.CTAClicks),
		TotalPayClicks: int64(stats.PayClicks),
		UniqueVisitors: int64(stats.UniqueVisitors),
	})
}
