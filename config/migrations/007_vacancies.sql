CREATE TABLE IF NOT EXISTS vacancies (
    id          BIGSERIAL    PRIMARY KEY,
    slug        VARCHAR(255) UNIQUE,
    user_id     BIGINT,
    region_id   BIGINT,
    district_id BIGINT,
    mahalla_id  BIGINT,
    adress      TEXT,
    name        VARCHAR(255),
    title       VARCHAR(500),
    text        TEXT,
    contact     VARCHAR(255),
    telegram    VARCHAR(255),
    price       BIGINT,
    views_count BIGINT       DEFAULT 0,
    is_active   BOOLEAN      DEFAULT TRUE,
    created_at  TIMESTAMPTZ  DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

-- Partial btree: faqat tirik yozuvlarni indekslaydi (hajm kichik, yozish arzon)
CREATE INDEX IF NOT EXISTS idx_vacancies_user_id         ON vacancies(user_id)                WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vacancies_price           ON vacancies(price)                  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vacancies_active_region   ON vacancies(region_id,   is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vacancies_active_district ON vacancies(district_id, is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vacancies_active_mahalla  ON vacancies(mahalla_id,  is_active) WHERE deleted_at IS NULL;

-- Trigram GIN: ILIKE '%...%' qidiruvi uchun (name / title / text)
CREATE INDEX IF NOT EXISTS idx_vacancies_name_trgm  ON vacancies USING GIN (name  gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vacancies_title_trgm ON vacancies USING GIN (title gin_trgm_ops) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vacancies_text_trgm  ON vacancies USING GIN (text  gin_trgm_ops) WHERE deleted_at IS NULL;
