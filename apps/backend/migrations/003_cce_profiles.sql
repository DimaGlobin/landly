-- +goose Up
-- +goose StatementBegin

ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS schema_version INTEGER NOT NULL DEFAULT 1;

CREATE TABLE IF NOT EXISTS brand_profiles (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    brand_name VARCHAR(255),
    brand_tone VARCHAR(50),
    font VARCHAR(100),
    style_preset VARCHAR(50),
    brand_colors JSONB,
    preferred_words JSONB,
    forbidden_words JSONB,
    guidelines JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(project_id)
);

CREATE INDEX IF NOT EXISTS idx_brand_profiles_project_id ON brand_profiles(project_id);

CREATE TABLE IF NOT EXISTS product_profiles (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    product_name VARCHAR(255),
    target_audience VARCHAR(255),
    goal VARCHAR(50),
    value_prop TEXT,
    differentiators JSONB,
    features JSONB,
    pricing JSONB,
    primary_link TEXT,
    payment_url TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(project_id)
);

CREATE INDEX IF NOT EXISTS idx_product_profiles_project_id ON product_profiles(project_id);

CREATE TABLE IF NOT EXISTS content_snippets (
    id UUID PRIMARY KEY,
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    label VARCHAR(120) NOT NULL,
    content TEXT NOT NULL,
    locale VARCHAR(16) NOT NULL DEFAULT 'ru',
    tags JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, label, locale)
);

CREATE INDEX IF NOT EXISTS idx_content_snippets_project_id ON content_snippets(project_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_content_snippets_project_id;
DROP TABLE IF EXISTS content_snippets;

DROP INDEX IF EXISTS idx_product_profiles_project_id;
DROP TABLE IF EXISTS product_profiles;

DROP INDEX IF EXISTS idx_brand_profiles_project_id;
DROP TABLE IF EXISTS brand_profiles;

ALTER TABLE projects
    DROP COLUMN IF EXISTS schema_version;

-- +goose StatementEnd
