package dto

import (
	"time"

	"github.com/google/uuid"
)

// Standard response wrapper
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Auth responses
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Project responses
type ProjectResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	Niche     string    `json:"niche"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ProjectsListResponse struct {
	Projects []ProjectResponse `json:"projects"`
	Total    int               `json:"total"`
}

type GenerationSessionResponse struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type PreviewResponse struct {
	Schema map[string]interface{} `json:"schema"`
}

type PublishResponse struct {
	Subdomain   string `json:"subdomain"`
	PublicURL   string `json:"public_url"`
	PublishedAt string `json:"published_at"`
}

// Analytics responses
type AnalyticsStatsResponse struct {
	ProjectID      uuid.UUID `json:"project_id"`
	TotalPageViews int64     `json:"total_pageviews"`
	TotalCTAClicks int64     `json:"total_cta_clicks"`
	TotalPayClicks int64     `json:"total_pay_clicks"`
	UniqueVisitors int64     `json:"unique_visitors"`
}
