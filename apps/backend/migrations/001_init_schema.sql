-- +goose Up
-- +goose StatementBegin

-- Пользователи
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
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

-- Страницы
CREATE TABLE IF NOT EXISTS pages (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    path VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    meta_json TEXT,
    sort INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, path)
);

CREATE INDEX IF NOT EXISTS idx_pages_project_id ON pages(project_id);

-- Блоки
CREATE TABLE IF NOT EXISTS blocks (
    id UUID PRIMARY KEY,
    page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    props_json TEXT NOT NULL,
    sort INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_blocks_page_id ON blocks(page_id);

-- Интеграции (payment URLs и т.д.)
CREATE TABLE IF NOT EXISTS integrations (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    value TEXT NOT NULL,
    meta_json TEXT,
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
    custom_domain VARCHAR(255),
    ssl_status VARCHAR(50) NOT NULL DEFAULT 'none',
    last_published_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_publish_targets_project_id ON publish_targets(project_id);
CREATE INDEX IF NOT EXISTS idx_publish_targets_subdomain ON publish_targets(subdomain);

-- Сессии генерации
CREATE TABLE IF NOT EXISTS generation_sessions (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    prompt TEXT NOT NULL,
    model VARCHAR(100) NOT NULL,
    output_json TEXT,
    duration_ms BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_generation_sessions_project_id ON generation_sessions(project_id);
CREATE INDEX IF NOT EXISTS idx_generation_sessions_created_at ON generation_sessions(created_at DESC);

-- Аналитические события
CREATE TABLE IF NOT EXISTS analytics_events (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL,
    path VARCHAR(255) NOT NULL,
    referrer TEXT,
    user_agent TEXT,
    ip_address VARCHAR(45),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_analytics_events_project_id ON analytics_events(project_id);
CREATE INDEX IF NOT EXISTS idx_analytics_events_created_at ON analytics_events(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_analytics_events_type ON analytics_events(event_type);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS analytics_events;
DROP TABLE IF EXISTS generation_sessions;
DROP TABLE IF EXISTS publish_targets;
DROP TABLE IF EXISTS integrations;
DROP TABLE IF EXISTS blocks;
DROP TABLE IF EXISTS pages;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;

-- +goose StatementEnd

