-- +goose Up
-- +goose StatementBegin

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

DROP INDEX IF EXISTS idx_schema_versions_project_created;
DROP INDEX IF EXISTS idx_schema_versions_created_at;
DROP INDEX IF EXISTS idx_schema_versions_project_id;
DROP TABLE IF EXISTS project_schema_versions;

-- +goose StatementEnd

