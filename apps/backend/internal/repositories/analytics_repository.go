package repositories

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	domain "github.com/landly/backend/internal/models"
	"github.com/landly/backend/internal/query"
)

// AnalyticsRepository интерфейс репозитория аналитики
type AnalyticsRepository interface {
	TrackEvent(ctx context.Context, event *domain.AnalyticsEvent) error
	GetStats(ctx context.Context, projectID uuid.UUID) (*domain.AnalyticsStats, error)
	GetEvents(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*domain.AnalyticsEvent, error)
}

// analyticsRepository реализация репозитория аналитики
type analyticsRepository struct {
	qb *query.Builder
}

// NewAnalyticsRepository создает новый репозиторий аналитики
func NewAnalyticsRepository(qb *query.Builder) AnalyticsRepository {
	return &analyticsRepository{qb: qb}
}

// TrackEvent отслеживает событие
func (r *analyticsRepository) TrackEvent(ctx context.Context, event *domain.AnalyticsEvent) error {
	query := r.qb.Insert("analytics_events").
		Columns("id", "project_id", "event_type", "path", "referrer", "user_agent", "ip_address", "created_at").
		Values(event.ID, event.ProjectID, event.EventType, event.Path, event.Referrer, event.UserAgent, event.IPAddress, event.CreatedAt)

	_, err := r.qb.Execute(query)
	return err
}

// GetStats получает статистику проекта
func (r *analyticsRepository) GetStats(ctx context.Context, projectID uuid.UUID) (*domain.AnalyticsStats, error) {
	// Используем raw SQL для сложных агрегаций
	var stats struct {
		TotalPageViews int64 `db:"total_page_views"`
		UniqueVisitors int64 `db:"unique_visitors"`
		CTAClicks      int64 `db:"cta_clicks"`
		PayClicks      int64 `db:"pay_clicks"`
	}

	// Диалект-агностичный запрос
	var sql string
	switch r.qb.GetDialect() {
	case query.PostgreSQL:
		sql = `
			SELECT 
				COUNT(CASE WHEN event_type = 'pageview' THEN 1 END) as total_page_views,
				COUNT(DISTINCT CASE WHEN event_type = 'pageview' THEN ip_address END) as unique_visitors,
				COUNT(CASE WHEN event_type = 'cta_click' THEN 1 END) as cta_clicks,
				COUNT(CASE WHEN event_type = 'pay_click' THEN 1 END) as pay_clicks
			FROM analytics_events 
			WHERE project_id = $1
		`
	case query.MySQL:
		sql = `
			SELECT 
				COUNT(CASE WHEN event_type = 'pageview' THEN 1 END) as total_page_views,
				COUNT(DISTINCT CASE WHEN event_type = 'pageview' THEN ip_address END) as unique_visitors,
				COUNT(CASE WHEN event_type = 'cta_click' THEN 1 END) as cta_clicks,
				COUNT(CASE WHEN event_type = 'pay_click' THEN 1 END) as pay_clicks
			FROM analytics_events 
			WHERE project_id = ?
		`
	case query.SQLite:
		sql = `
			SELECT 
				COUNT(CASE WHEN event_type = 'pageview' THEN 1 END) as total_page_views,
				COUNT(DISTINCT CASE WHEN event_type = 'pageview' THEN ip_address END) as unique_visitors,
				COUNT(CASE WHEN event_type = 'cta_click' THEN 1 END) as cta_clicks,
				COUNT(CASE WHEN event_type = 'pay_click' THEN 1 END) as pay_clicks
			FROM analytics_events 
			WHERE project_id = ?
		`
	}

	row := r.qb.GetDB().QueryRowContext(ctx, sql, projectID)
	err := row.Scan(&stats.TotalPageViews, &stats.UniqueVisitors, &stats.CTAClicks, &stats.PayClicks)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}

	return &domain.AnalyticsStats{
		TotalPageViews: int(stats.TotalPageViews),
		UniqueVisitors: int(stats.UniqueVisitors),
		CTAClicks:      int(stats.CTAClicks),
		PayClicks:      int(stats.PayClicks),
	}, nil
}

// GetEvents получает события аналитики
func (r *analyticsRepository) GetEvents(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*domain.AnalyticsEvent, error) {
	query := r.qb.Select("id", "project_id", "event_type", "path", "referrer", "user_agent", "ip_address", "created_at").
		From("analytics_events").
		Where(squirrel.Eq{"project_id": projectID}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	rows, err := r.qb.Query(query)
	if err != nil {
		return nil, domain.ErrInternal.WithError(err)
	}
	defer rows.Close()

	var events []*domain.AnalyticsEvent
	for rows.Next() {
		var event domain.AnalyticsEvent
		err := rows.Scan(&event.ID, &event.ProjectID, &event.EventType, &event.Path, &event.Referrer, &event.UserAgent, &event.IPAddress, &event.CreatedAt)
		if err != nil {
			return nil, domain.ErrInternal.WithError(err)
		}
		events = append(events, &event)
	}

	return events, nil
}
