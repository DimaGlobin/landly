package services

import (
	"context"

	"github.com/google/uuid"
	domain "github.com/landly/backend/internal/models"
)

// AnalyticsService сервис для аналитики
type AnalyticsService struct {
	projectRepo   domain.ProjectRepository
	analyticsRepo domain.AnalyticsRepository
}

// NewAnalyticsService создаёт новый analytics service
func NewAnalyticsService(
	projectRepo domain.ProjectRepository,
	analyticsRepo domain.AnalyticsRepository,
) *AnalyticsService {
	return &AnalyticsService{
		projectRepo:   projectRepo,
		analyticsRepo: analyticsRepo,
	}
}

// TrackEvent отслеживает событие (новый интерфейс)
func (s *AnalyticsService) TrackEvent(ctx context.Context, req *domain.TrackEventRequest) error {
	// Создаем событие
	event := &domain.AnalyticsEvent{
		ID:        uuid.New(),
		EventType: req.EventType,
		Path:      req.Path,
		Referrer:  req.Referrer,
	}

	if err := s.analyticsRepo.TrackEvent(ctx, event); err != nil {
		// Не критично, можем игнорировать ошибки трекинга
		return nil
	}

	return nil
}

// GetProjectAnalytics получает аналитику проекта
func (s *AnalyticsService) GetProjectAnalytics(ctx context.Context, userID, projectID string) (*domain.ProjectAnalytics, error) {
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid project ID")
	}

	// Проверка доступа к проекту
	project, err := s.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, domain.ErrNotFound.WithMessage("project not found")
	}

	if project.UserID.String() != userID {
		return nil, domain.ErrForbidden.WithMessage("access denied")
	}

	stats, err := s.analyticsRepo.GetStats(ctx, projectUUID)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	return &domain.ProjectAnalytics{
		TotalPageViews: stats.TotalPageViews,
		UniqueVisitors: stats.UniqueVisitors,
		CTAClicks:      stats.CTAClicks,
		PayClicks:      stats.PayClicks,
	}, nil
}

// GetSiteAnalytics получает аналитику сайта
func (s *AnalyticsService) GetSiteAnalytics(ctx context.Context, userID, targetID string) (*domain.SiteAnalytics, error) {
	targetUUID, err := uuid.Parse(targetID)
	if err != nil {
		return nil, domain.ErrBadRequest.WithMessage("invalid target ID")
	}

	stats, err := s.analyticsRepo.GetStats(ctx, targetUUID)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	return &domain.SiteAnalytics{
		ProjectAnalytics: domain.ProjectAnalytics{
			TotalPageViews: stats.TotalPageViews,
			UniqueVisitors: stats.UniqueVisitors,
			CTAClicks:      stats.CTAClicks,
			PayClicks:      stats.PayClicks,
		},
		LastPublishedAt: nil, // TODO: получить из publish target
	}, nil
}

// GetProjectStats получает статистику проекта (старый интерфейс)
func (s *AnalyticsService) GetProjectStats(ctx context.Context, userID, projectID uuid.UUID) (*domain.AnalyticsStats, error) {
	// Проверка доступа к проекту
	project, err := s.projectRepo.GetByID(ctx, projectID.String())
	if err != nil {
		return nil, domain.ErrNotFound.WithMessage("project not found")
	}

	if project.UserID != userID {
		return nil, domain.ErrForbidden.WithMessage("access denied")
	}

	// Получаем статистику
	stats, err := s.analyticsRepo.GetStats(ctx, projectID)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	return stats, nil
}

// GetProjectEvents получает события проекта
func (s *AnalyticsService) GetProjectEvents(ctx context.Context, userID, projectID uuid.UUID, limit, offset int) ([]*domain.AnalyticsEvent, error) {
	// Проверка доступа к проекту
	project, err := s.projectRepo.GetByID(ctx, projectID.String())
	if err != nil {
		return nil, domain.ErrNotFound.WithMessage("project not found")
	}

	if project.UserID != userID {
		return nil, domain.ErrForbidden.WithMessage("access denied")
	}

	// Получаем события
	events, err := s.analyticsRepo.GetEvents(ctx, projectID, limit, offset)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	return events, nil
}
