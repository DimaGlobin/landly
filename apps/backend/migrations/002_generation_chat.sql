-- +goose Up
-- +goose StatementBegin

ALTER TABLE generation_sessions
    ADD COLUMN IF NOT EXISTS status VARCHAR(50) NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS schema_json TEXT,
    ADD COLUMN IF NOT EXISTS completed_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP NOT NULL DEFAULT NOW();

ALTER TABLE generation_sessions
    DROP COLUMN IF EXISTS output_json,
    DROP COLUMN IF EXISTS duration_ms;

CREATE TABLE IF NOT EXISTS generation_messages (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES generation_sessions(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL,
    content TEXT NOT NULL,
    metadata TEXT,
    tokens_used INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_generation_messages_session_created_at
    ON generation_messages(session_id, created_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_generation_messages_session_created_at;
DROP TABLE IF EXISTS generation_messages;

ALTER TABLE generation_sessions
    ADD COLUMN IF NOT EXISTS output_json TEXT,
    ADD COLUMN IF NOT EXISTS duration_ms BIGINT;

ALTER TABLE generation_sessions
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS completed_at,
    DROP COLUMN IF EXISTS schema_json,
    DROP COLUMN IF EXISTS status;

-- +goose StatementEnd

