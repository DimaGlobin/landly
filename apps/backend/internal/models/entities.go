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
	ID            uuid.UUID `db:"id" json:"id"`
	UserID        uuid.UUID `db:"user_id" json:"user_id"`
	Name          string    `db:"name" json:"name"`
	Niche         string    `db:"niche" json:"niche"`
	SchemaJSON    string    `db:"schema_json" json:"schema_json"`
	SchemaVersion int       `db:"schema_version" json:"schema_version"`
	Status        string    `db:"status" json:"status"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
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

// GenerationMessage представляет сообщение в рамках сессии генерации
type GenerationMessage struct {
	ID         uuid.UUID `db:"id" json:"id"`
	SessionID  uuid.UUID `db:"session_id" json:"session_id"`
	Role       string    `db:"role" json:"role"`
	Content    string    `db:"content" json:"content"`
	Metadata   string    `db:"metadata" json:"metadata"`
	TokensUsed int       `db:"tokens_used" json:"tokens_used"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
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

// BrandProfile описывает брендовые настройки проекта
type BrandProfile struct {
	ID             uuid.UUID         `db:"id" json:"id"`
	ProjectID      uuid.UUID         `db:"project_id" json:"project_id"`
	BrandName      string            `db:"brand_name" json:"brand_name"`
	BrandTone      string            `db:"brand_tone" json:"brand_tone"`
	Font           string            `db:"font" json:"font"`
	StylePreset    string            `db:"style_preset" json:"style_preset"`
	BrandColors    []string          `db:"brand_colors" json:"brand_colors"`
	PreferredWords []string          `db:"preferred_words" json:"preferred_words"`
	ForbiddenWords []string          `db:"forbidden_words" json:"forbidden_words"`
	Guidelines     map[string]string `db:"guidelines" json:"guidelines"`
	CreatedAt      time.Time         `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time         `db:"updated_at" json:"updated_at"`
}

// ProductFeature описывает фичу продукта для промпта
type ProductFeature struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// ProductPlan описывает тариф
type ProductPlan struct {
	Name       string   `json:"name"`
	Price      string   `json:"price"`
	Currency   string   `json:"currency"`
	Period     string   `json:"period"`
	Features   []string `json:"features"`
	ButtonText string   `json:"button_text"`
	URL        string   `json:"url"`
	Featured   bool     `json:"featured"`
}

// ProductProfile содержит продуктовый контент проекта
type ProductProfile struct {
	ID              uuid.UUID        `db:"id" json:"id"`
	ProjectID       uuid.UUID        `db:"project_id" json:"project_id"`
	ProductName     string           `db:"product_name" json:"product_name"`
	TargetAudience  string           `db:"target_audience" json:"target_audience"`
	Goal            string           `db:"goal" json:"goal"`
	ValueProp       string           `db:"value_prop" json:"value_prop"`
	Differentiators []string         `db:"differentiators" json:"differentiators"`
	Features        []ProductFeature `db:"features" json:"features"`
	Pricing         []ProductPlan    `db:"pricing" json:"pricing"`
	PrimaryLink     string           `db:"primary_link" json:"primary_link"`
	PaymentURL      string           `db:"payment_url" json:"payment_url"`
	CreatedAt       time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time        `db:"updated_at" json:"updated_at"`
}

// ContentSnippet представляет текстовый фрагмент, который нужно включить в промпт
type ContentSnippet struct {
	ID        uuid.UUID `db:"id" json:"id"`
	ProjectID uuid.UUID `db:"project_id" json:"project_id"`
	Label     string    `db:"label" json:"label"`
	Content   string    `db:"content" json:"content"`
	Locale    string    `db:"locale" json:"locale"`
	Tags      []string  `db:"tags" json:"tags"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
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

	MessageRoleUser      = "user"
	MessageRoleAssistant = "assistant"
	MessageRoleSystem    = "system"

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
		ID:            uuid.New(),
		UserID:        userID,
		Name:          name,
		Niche:         niche,
		SchemaVersion: 1,
		Status:        ProjectStatusDraft,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
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
