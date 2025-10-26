package dto

// Auth requests
type SignUpRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type SignInRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Project requests
type CreateProjectRequest struct {
	Name  string `json:"name" binding:"required"`
	Niche string `json:"niche" binding:"required"`
}

type UpdateProjectRequest struct {
	Name  string `json:"name"`
	Niche string `json:"niche"`
}

// Generate requests
type GenerateRequest struct {
	Prompt     string `json:"prompt" binding:"required"`
	PaymentURL string `json:"payment_url"`
}

// Analytics requests
type TrackEventRequest struct {
	EventType string `json:"event_type" binding:"required"`
	Path      string `json:"path" binding:"required"`
	Referrer  string `json:"referrer"`
}
