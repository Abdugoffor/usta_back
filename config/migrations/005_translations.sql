CREATE TABLE IF NOT EXISTS translations (
    id         BIGSERIAL    PRIMARY KEY,
    slug       VARCHAR(150) NOT NULL UNIQUE,
    name       JSONB        NOT NULL,
    is_active  BOOLEAN      DEFAULT TRUE,
    created_at TIMESTAMPTZ  DEFAULT NOW(),
    updated_at TIMESTAMPTZ  DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_translations_slug  ON translations(slug);
CREATE INDEX IF NOT EXISTS idx_translations_is_active  ON translations(is_active);
CREATE INDEX IF NOT EXISTS idx_translations_deleted_at ON translations(deleted_at);
