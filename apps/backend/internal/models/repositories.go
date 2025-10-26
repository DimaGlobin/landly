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
	GetByUserID(ctx context.Context, userID string) ([]*Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id string) error
	UpdateSchema(ctx context.Context, projectID string, schemaJSON string) error
}

// PageRepository интерфейс репозитория страниц
type PageRepository interface {
	Create(ctx context.Context, page *Page) error
	GetByID(ctx context.Context, id uuid.UUID) (*Page, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*Page, error)
	Update(ctx context.Context, page *Page) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByProjectID(ctx context.Context, projectID uuid.UUID) error
}

// BlockRepository интерфейс репозитория блоков
type BlockRepository interface {
	Create(ctx context.Context, block *Block) error
	GetByID(ctx context.Context, id uuid.UUID) (*Block, error)
	GetByPageID(ctx context.Context, pageID uuid.UUID) ([]*Block, error)
	Update(ctx context.Context, block *Block) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByPageID(ctx context.Context, pageID uuid.UUID) error
}

// IntegrationRepository интерфейс репозитория интеграций
type IntegrationRepository interface {
	Create(ctx context.Context, integration *Integration) error
	GetByID(ctx context.Context, id string) (*Integration, error)
	GetByProjectID(ctx context.Context, projectID string) ([]*Integration, error)
	GetByProjectIDAndType(ctx context.Context, projectID string, integrationType IntegrationType) (*Integration, error)
	Update(ctx context.Context, integration *Integration) error
	Delete(ctx context.Context, id string) error
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

// GenerationSessionRepository интерфейс репозитория сессий генерации
type GenerationSessionRepository interface {
	Create(ctx context.Context, session *GenerationSession) error
	GetByID(ctx context.Context, id string) (*GenerationSession, error)
	GetByProjectID(ctx context.Context, projectID string) ([]*GenerationSession, error)
	Update(ctx context.Context, session *GenerationSession) error
	Delete(ctx context.Context, id string) error
}

// AnalyticsRepository интерфейс репозитория аналитики
type AnalyticsRepository interface {
	TrackEvent(ctx context.Context, event *AnalyticsEvent) error
	GetStats(ctx context.Context, projectID uuid.UUID) (*AnalyticsStats, error)
	GetEvents(ctx context.Context, projectID uuid.UUID, limit, offset int) ([]*AnalyticsEvent, error)
}
