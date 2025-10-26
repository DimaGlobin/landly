package domain

import (
	"time"

	"github.com/google/uuid"
)

// User представляет пользователя системы
type User struct {
	ID           uuid.UUID `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"password_hash"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// Project представляет проект пользователя
type Project struct {
	ID         uuid.UUID `db:"id" json:"id"`
	UserID     uuid.UUID `db:"user_id" json:"user_id"`
	Name       string    `db:"name" json:"name"`
	Niche      string    `db:"niche" json:"niche"`
	SchemaJSON string    `db:"schema_json" json:"schema_json"`
	Status     string    `db:"status" json:"status"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

// GenerationSession представляет сессию генерации
type GenerationSession struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	ProjectID   uuid.UUID  `db:"project_id" json:"project_id"`
	Prompt      string     `db:"prompt" json:"prompt"`
	Model       string     `db:"model" json:"model"`
	Status      string     `db:"status" json:"status"`
	SchemaJSON  string     `db:"schema_json" json:"schema_json"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

// PublishTarget представляет цель публикации
type PublishTarget struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	ProjectID       uuid.UUID  `db:"project_id" json:"project_id"`
	Subdomain       string     `db:"subdomain" json:"subdomain"`
	Status          string     `db:"status" json:"status"`
	LastPublishedAt *time.Time `db:"last_published_at" json:"last_published_at"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

// AnalyticsEvent представляет событие аналитики
type AnalyticsEvent struct {
	ID        uuid.UUID `db:"id" json:"id"`
	ProjectID uuid.UUID `db:"project_id" json:"project_id"`
	EventType string    `db:"event_type" json:"event_type"`
	Path      string    `db:"path" json:"path"`
	Referrer  string    `db:"referrer" json:"referrer"`
	UserAgent string    `db:"user_agent" json:"user_agent"`
	IPAddress string    `db:"ip_address" json:"ip_address"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// Integration представляет интеграцию
type Integration struct {
	ID        uuid.UUID `db:"id" json:"id"`
	ProjectID uuid.UUID `db:"project_id" json:"project_id"`
	Type      string    `db:"type" json:"type"`
	Config    string    `db:"config" json:"config"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// Константы статусов
const (
	ProjectStatusDraft     = "draft"
	ProjectStatusGenerated = "generated"
	ProjectStatusPublished = "published"

	GenerationStatusPending   = "pending"
	GenerationStatusCompleted = "completed"
	GenerationStatusFailed    = "failed"

	PublishStatusDraft     = "draft"
	PublishStatusPublished = "published"
	PublishStatusFailed    = "failed"

	IntegrationTypeStripe = "stripe"
	IntegrationTypePayPal = "paypal"
)

// IntegrationType тип интеграции
type IntegrationType string

// AnalyticsStats статистика аналитики
type AnalyticsStats struct {
	TotalPageViews int `json:"total_page_views"`
	UniqueVisitors int `json:"unique_visitors"`
	CTAClicks      int `json:"cta_clicks"`
	PayClicks      int `json:"pay_clicks"`
}

// ProjectAnalytics аналитика проекта
type ProjectAnalytics struct {
	TotalPageViews int `json:"total_page_views"`
	UniqueVisitors int `json:"unique_visitors"`
	CTAClicks      int `json:"cta_clicks"`
	PayClicks      int `json:"pay_clicks"`
}

// SiteAnalytics аналитика сайта
type SiteAnalytics struct {
	ProjectAnalytics
	LastPublishedAt *time.Time `json:"last_published_at"`
}

// PageSchema схема страницы
type PageSchema struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Blocks      []BlockData `json:"blocks"`
}

// BlockData данные блока
type BlockData struct {
	Type  BlockType              `json:"type"`
	Props map[string]interface{} `json:"props"`
}

// GenerationResult результат генерации
type GenerationResult struct {
	Schema *PageSchema `json:"schema"`
	HTML   string      `json:"html"`
}

// Конструкторы

// NewProject создаёт новый проект
func NewProject(userID uuid.UUID, name, niche string) *Project {
	return &Project{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		Niche:     niche,
		Status:    ProjectStatusDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewGenerationSession создаёт новую сессию генерации
func NewGenerationSession(projectID uuid.UUID, prompt, model string) *GenerationSession {
	return &GenerationSession{
		ID:        uuid.New(),
		ProjectID: projectID,
		Prompt:    prompt,
		Model:     model,
		Status:    GenerationStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewPublishTarget создаёт новую цель публикации
func NewPublishTarget(projectID uuid.UUID, subdomain string) *PublishTarget {
	return &PublishTarget{
		ID:        uuid.New(),
		ProjectID: projectID,
		Subdomain: subdomain,
		Status:    PublishStatusDraft,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewAnalyticsEvent создаёт новое событие аналитики
func NewAnalyticsEvent(projectID uuid.UUID, eventType, path, referrer, userAgent, ipAddress string) *AnalyticsEvent {
	return &AnalyticsEvent{
		ID:        uuid.New(),
		ProjectID: projectID,
		EventType: eventType,
		Path:      path,
		Referrer:  referrer,
		UserAgent: userAgent,
		IPAddress: ipAddress,
		CreatedAt: time.Now(),
	}
}

// NewIntegration создаёт новую интеграцию
func NewIntegration(projectID uuid.UUID, integrationType IntegrationType, config string) *Integration {
	return &Integration{
		ID:        uuid.New(),
		ProjectID: projectID,
		Type:      string(integrationType),
		Config:    config,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
