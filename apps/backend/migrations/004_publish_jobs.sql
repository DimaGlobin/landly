-- +goose Up
-- +goose StatementBegin

-- Таблица для управления задачами публикации
CREATE TABLE IF NOT EXISTS publish_jobs (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    publish_target_id UUID REFERENCES publish_targets(id) ON DELETE SET NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    release_id UUID,
    error_message TEXT,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_publish_jobs_project_id ON publish_jobs(project_id);
CREATE INDEX IF NOT EXISTS idx_publish_jobs_status ON publish_jobs(status);
CREATE INDEX IF NOT EXISTS idx_publish_jobs_release_id ON publish_jobs(release_id);
CREATE INDEX IF NOT EXISTS idx_publish_jobs_publish_target_id ON publish_jobs(publish_target_id);

-- Добавляем поле current_release_id в publish_targets для атомарного переключения
-- (если колонка уже существует в 001_init_schema.sql, это не вызовет ошибку благодаря IF NOT EXISTS)
ALTER TABLE publish_targets ADD COLUMN IF NOT EXISTS current_release_id UUID;

-- Индекс для быстрого поиска по current_release_id
CREATE INDEX IF NOT EXISTS idx_publish_targets_current_release_id ON publish_targets(current_release_id);

-- +goose StatementEnd

