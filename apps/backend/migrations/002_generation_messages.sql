-- +goose Up
-- +goose StatementBegin

-- Сообщения чата генерации
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

-- +goose StatementEnd

