CREATE TABLE IF NOT EXISTS countries (
    id         BIGSERIAL    PRIMARY KEY,
    parent_id  BIGINT,
    name       JSONB        DEFAULT '{}'::jsonb,
    is_active  BOOLEAN      DEFAULT TRUE,
    created_at TIMESTAMPTZ  DEFAULT NOW(),
    updated_at TIMESTAMPTZ  DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_countries_parent_id  ON countries(parent_id);
CREATE INDEX IF NOT EXISTS idx_countries_is_active  ON countries(is_active);
CREATE INDEX IF NOT EXISTS idx_countries_deleted_at ON countries(deleted_at);
