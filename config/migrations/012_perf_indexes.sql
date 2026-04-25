-- Tezkor List/Count uchun yetishmayotgan indekslar.
-- Hammasi partial (WHERE deleted_at IS NULL) va trigram (ILIKE '%...%' uchun).

-- ── users ─────────────────────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_users_role_active     ON users(role, is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_created_at      ON users(created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_full_name_trgm  ON users USING GIN (full_name gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_phone_trgm      ON users USING GIN (phone     gin_trgm_ops) WHERE deleted_at IS NULL;

-- ── languages ─────────────────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_languages_name_trgm        ON languages USING GIN (name        gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_languages_description_trgm ON languages USING GIN (description gin_trgm_ops) WHERE deleted_at IS NULL;

-- ── categories ────────────────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_categories_active_id   ON categories(id DESC)         WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_categories_name_trgm   ON categories USING GIN ((name::text) gin_trgm_ops) WHERE deleted_at IS NULL;

-- ── countries ─────────────────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_countries_parent_active ON countries(parent_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_countries_name_trgm     ON countries USING GIN ((name::text) gin_trgm_ops) WHERE deleted_at IS NULL;

-- ── translations ──────────────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_translations_active_id  ON translations(id DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_translations_slug_trgm  ON translations USING GIN (slug gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_translations_name_trgm  ON translations USING GIN ((name::text) gin_trgm_ops) WHERE deleted_at IS NULL;

-- ── vacancies / resumes — keyset uchun id DESC partial ────────────────────
CREATE INDEX IF NOT EXISTS idx_vacancies_active_id ON vacancies(id DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_resumes_active_id   ON resumes(id DESC)   WHERE deleted_at IS NULL;

-- ── comments ──────────────────────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_comments_root_vacancy ON comments(vakansiya_id, id DESC) WHERE deleted_at IS NULL AND parent_id IS NULL;
CREATE INDEX IF NOT EXISTS idx_comments_root_resume  ON comments(resume_id,    id DESC) WHERE deleted_at IS NULL AND parent_id IS NULL;
