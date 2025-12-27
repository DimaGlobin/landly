-- +goose Up
-- +goose StatementBegin

-- Оптимизация схемы: удаление неиспользуемых таблиц и добавление constraints
-- Эта миграция безопасна для применения на существующих БД

-- Удаление неиспользуемых таблиц (если они существуют)
DROP TABLE IF EXISTS blocks;
DROP TABLE IF EXISTS pages;

-- Исправление несоответствия в integrations (для старых БД)
-- Переименование value -> config если колонка value существует
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'integrations' AND column_name = 'value'
    ) THEN
        ALTER TABLE integrations RENAME COLUMN value TO config;
    END IF;
END $$;

-- Добавление UNIQUE constraint на publish_targets(project_id) для 1:1 отношения
-- Удаление дубликатов перед добавлением constraint
DO $$
DECLARE
    dup_record RECORD;
BEGIN
    -- Удаляем дубликаты, оставляя самую свежую запись
    FOR dup_record IN 
        SELECT project_id, array_agg(id ORDER BY updated_at DESC) as ids
        FROM publish_targets
        GROUP BY project_id
        HAVING COUNT(*) > 1
    LOOP
        DELETE FROM publish_targets
        WHERE project_id = dup_record.project_id
          AND id != dup_record.ids[1];
    END LOOP;
    
    -- Добавляем constraint если его еще нет
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint 
        WHERE conname = 'publish_targets_project_id_unique'
    ) THEN
        ALTER TABLE publish_targets
        ADD CONSTRAINT publish_targets_project_id_unique UNIQUE (project_id);
    END IF;
END $$;

-- Исправление типов данных для существующих БД
-- password_hash: VARCHAR(255) -> TEXT
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'users' 
        AND column_name = 'password_hash' 
        AND data_type = 'character varying'
        AND character_maximum_length = 255
    ) THEN
        ALTER TABLE users ALTER COLUMN password_hash TYPE TEXT;
    END IF;
END $$;

-- analytics_events.path: VARCHAR(255) -> VARCHAR(500)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'analytics_events' 
        AND column_name = 'path' 
        AND character_maximum_length = 255
    ) THEN
        ALTER TABLE analytics_events ALTER COLUMN path TYPE VARCHAR(500);
    END IF;
END $$;

-- Добавление недостающих индексов (если их еще нет)
CREATE INDEX IF NOT EXISTS idx_publish_targets_status ON publish_targets(status);
CREATE INDEX IF NOT EXISTS idx_generation_sessions_project_created 
    ON generation_sessions(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_analytics_events_project_created 
    ON analytics_events(project_id, created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Откат изменений (для тестирования, в production обычно не используется)

DROP INDEX IF EXISTS idx_analytics_events_project_created;
DROP INDEX IF EXISTS idx_generation_sessions_project_created;
DROP INDEX IF EXISTS idx_publish_targets_status;

ALTER TABLE analytics_events ALTER COLUMN path TYPE VARCHAR(255);
ALTER TABLE users ALTER COLUMN password_hash TYPE VARCHAR(255);

ALTER TABLE publish_targets DROP CONSTRAINT IF EXISTS publish_targets_project_id_unique;

ALTER TABLE integrations RENAME COLUMN config TO value;

-- +goose StatementEnd

