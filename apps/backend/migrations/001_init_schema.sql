-- +goose Up
-- +goose StatementBegin

-- Пользователи
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Проекты
CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    niche VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    schema_json TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_projects_user_id ON projects(user_id);
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);

-- Интеграции (payment URLs и т.д.)
CREATE TABLE IF NOT EXISTS integrations (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    config TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, type)
);

CREATE INDEX IF NOT EXISTS idx_integrations_project_id ON integrations(project_id);

-- Цели публикации
CREATE TABLE IF NOT EXISTS publish_targets (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    subdomain VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    custom_domain VARCHAR(255),
    ssl_status VARCHAR(50) NOT NULL DEFAULT 'none',
    current_release_id UUID,
    last_published_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT publish_targets_project_id_unique UNIQUE (project_id)
);

CREATE INDEX IF NOT EXISTS idx_publish_targets_project_id ON publish_targets(project_id);
CREATE INDEX IF NOT EXISTS idx_publish_targets_subdomain ON publish_targets(subdomain);
CREATE INDEX IF NOT EXISTS idx_publish_targets_status ON publish_targets(status);

-- Сессии генерации (с полями для чата)
CREATE TABLE IF NOT EXISTS generation_sessions (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    prompt TEXT NOT NULL,
    model VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    schema_json TEXT,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_generation_sessions_project_id ON generation_sessions(project_id);
CREATE INDEX IF NOT EXISTS idx_generation_sessions_created_at ON generation_sessions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_generation_sessions_project_created ON generation_sessions(project_id, created_at DESC);

-- Аналитические события
CREATE TABLE IF NOT EXISTS analytics_events (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    path VARCHAR(500) NOT NULL,
    referrer TEXT,
    user_agent TEXT,
    ip_address VARCHAR(45),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_analytics_events_project_id ON analytics_events(project_id);
CREATE INDEX IF NOT EXISTS idx_analytics_events_created_at ON analytics_events(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_analytics_events_type ON analytics_events(event_type);
CREATE INDEX IF NOT EXISTS idx_analytics_events_project_created ON analytics_events(project_id, created_at DESC);

-- Версионирование схем проектов
CREATE TABLE IF NOT EXISTS project_schema_versions (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    schema_json TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    source VARCHAR(50) NOT NULL DEFAULT 'chat',
    tokens_used INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_schema_versions_project_id ON project_schema_versions(project_id);
CREATE INDEX IF NOT EXISTS idx_schema_versions_created_at ON project_schema_versions(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_schema_versions_project_created ON project_schema_versions(project_id, created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS project_schema_versions;
DROP TABLE IF EXISTS analytics_events;
DROP TABLE IF EXISTS generation_sessions;
DROP TABLE IF EXISTS publish_targets;
DROP TABLE IF EXISTS integrations;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;

-- +goose StatementEnd
