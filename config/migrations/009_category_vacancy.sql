CREATE TABLE IF NOT EXISTS category_vacancy (
    id           BIGSERIAL   PRIMARY KEY,
    categorya_id BIGINT,
    vacancy_id   BIGINT,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(categorya_id, vacancy_id)
);

CREATE INDEX IF NOT EXISTS idx_cat_vacancy_categorya ON category_vacancy(categorya_id);
CREATE INDEX IF NOT EXISTS idx_cat_vacancy_vacancy   ON category_vacancy(vacancy_id);
