CREATE TABLE IF NOT EXISTS category_resume (
    id           BIGSERIAL   PRIMARY KEY,
    categorya_id BIGINT,
    resume_id    BIGINT,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(categorya_id, resume_id)
);

CREATE INDEX IF NOT EXISTS idx_cat_resume_categorya ON category_resume(categorya_id);
CREATE INDEX IF NOT EXISTS idx_cat_resume_resume    ON category_resume(resume_id);
