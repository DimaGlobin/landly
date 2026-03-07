package domain

import (
	"context"

	"github.com/google/uuid"
)

// UserRepository интерфейс репозитория пользователей
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
}

// ProjectRepository интерфейс репозитория проектов
type ProjectRepository interface {
	Create(ctx context.Context, project *Project) error
	GetByID(ctx context.Context, id string) (*Project, error)
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id string) error
	UpdateSchema(ctx context.Context, projectID string, schemaJSON string) error
}

// PublishTargetRepository интерфейс репозитория целей публикации
type PublishTargetRepository interface {
	Create(ctx context.Context, target *PublishTarget) error
	GetByID(ctx context.Context, id string) (*PublishTarget, error)
	GetByProjectID(ctx context.Context, projectID string) (*PublishTarget, error)
	GetBySubdomain(ctx context.Context, subdomain string) (*PublishTarget, error)
	Update(ctx context.Context, target *PublishTarget) error
	Delete(ctx context.Context, id string) error
}

// PublishJobRepository интерфейс репозитория задач публикации
type PublishJobRepository interface {
	Create(ctx context.Context, job *PublishJob) error
	GetByID(ctx context.Context, id string) (*PublishJob, error)
	GetByProjectID(ctx context.Context, projectID string) ([]*PublishJob, error)
	GetByStatus(ctx context.Context, status string) ([]*PublishJob, error)
	GetPendingByProjectID(ctx context.Context, projectID string) (*PublishJob, error)
	Update(ctx context.Context, job *PublishJob) error
	Delete(ctx context.Context, id string) error
}

// IntegrationRepository интерфейс репозитория интеграций
type IntegrationRepository interface {
	Create(ctx context.Context, integration *Integration) error
	GetByID(ctx context.Context, id uuid.UUID) (*Integration, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*Integration, error)
	GetByProjectIDAndType(ctx context.Context, projectID uuid.UUID, integrationType string) (*Integration, error)
	Update(ctx context.Context, integration *Integration) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// AnalyticsRepository интерфейс репозитория аналитики
type AnalyticsRepository interface {
	TrackEvent(ctx context.Context, event *AnalyticsEvent) error
	GetStats(ctx context.Context, projectID uuid.UUID) (*AnalyticsStats, error)
	GetEvents(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*AnalyticsEvent, error)
}

// GenerationSessionRepository интерфейс репозитория сессий генерации
type GenerationSessionRepository interface {
	Create(ctx context.Context, session *GenerationSession) error
	GetByID(ctx context.Context, id string) (*GenerationSession, error)
	GetByProjectID(ctx context.Context, projectID string) ([]*GenerationSession, error)
	GetLatestByProjectID(ctx context.Context, projectID string) (*GenerationSession, error)
	Update(ctx context.Context, session *GenerationSession) error
	Delete(ctx context.Context, id string) error
}

// GenerationMessageRepository интерфейс репозитория сообщений генерации
type GenerationMessageRepository interface {
	Create(ctx context.Context, message *GenerationMessage) error
	ListBySession(ctx context.Context, sessionID string) ([]*GenerationMessage, error)
	DeleteBySession(ctx context.Context, sessionID string) error
}

// SchemaVersionRepository интерфейс репозитория версий схем
type SchemaVersionRepository interface {
	Create(ctx context.Context, version *SchemaVersion) error
	GetByID(ctx context.Context, id string) (*SchemaVersion, error)
	GetByProjectID(ctx context.Context, projectID string, limit int) ([]*SchemaVersion, error)
	Delete(ctx context.Context, id string) error
}
