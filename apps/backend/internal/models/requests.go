package domain

import (
	"time"
)

// Auth requests and responses
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Project requests and responses
type CreateProjectRequest struct {
	Name  string `json:"name" binding:"required"`
	Niche string `json:"niche" binding:"required"`
}

type UpdateProjectRequest struct {
	Name  string `json:"name"`
	Niche string `json:"niche"`
}

// Generate requests and responses
type GenerateRequest struct {
	Prompt     string `json:"prompt" binding:"required"`
	PaymentURL string `json:"payment_url"`
}

// Publish requests and responses
type PublishRequest struct {
	Domain string `json:"domain"`
	Path   string `json:"path"`
}

// Analytics requests and responses
type TrackEventRequest struct {
	EventType string `json:"event_type" binding:"required"`
	Path      string `json:"path" binding:"required"`
	Referrer  string `json:"referrer"`
}
