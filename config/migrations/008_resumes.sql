CREATE TABLE IF NOT EXISTS resumes (
    id              BIGSERIAL    PRIMARY KEY,
    slug            VARCHAR(255) UNIQUE,
    user_id         BIGINT,
    region_id       BIGINT,
    district_id     BIGINT,
    mahalla_id      BIGINT,
    adress          TEXT,
    name            VARCHAR(255),
    photo           TEXT,
    title           VARCHAR(500),
    text            TEXT,
    contact         VARCHAR(255),
    telegram        VARCHAR(255),
    price           BIGINT,
    experience_year INT,
    skills          TEXT,
    views_count     BIGINT       DEFAULT 0,
    is_active       BOOLEAN      DEFAULT TRUE,
    created_at      TIMESTAMPTZ  DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

-- Partial btree: faqat tirik yozuvlarni indekslaydi (hajm kichik, yozish arzon)
CREATE INDEX IF NOT EXISTS idx_resumes_user_id         ON resumes(user_id)                 WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_resumes_price           ON resumes(price)                   WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_resumes_experience      ON resumes(experience_year)         WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_resumes_active_region   ON resumes(region_id,   is_active)  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_resumes_active_district ON resumes(district_id, is_active)  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_resumes_active_mahalla  ON resumes(mahalla_id,  is_active)  WHERE deleted_at IS NULL;

-- Trigram GIN: ILIKE '%...%' qidiruvi uchun (name / title / skills / text)
CREATE INDEX IF NOT EXISTS idx_resumes_name_trgm   ON resumes USING GIN (name   gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_resumes_title_trgm  ON resumes USING GIN (title  gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_resumes_skills_trgm ON resumes USING GIN (skills gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_resumes_text_trgm   ON resumes USING GIN (text   gin_trgm_ops) WHERE deleted_at IS NULL;
