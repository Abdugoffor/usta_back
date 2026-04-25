CREATE TABLE IF NOT EXISTS comments (
    id           BIGSERIAL   PRIMARY KEY,
    parent_id    BIGINT,
    user_id      BIGINT,
    vakansiya_id BIGINT,
    resume_id    BIGINT,
    type         VARCHAR(50),
    text         TEXT,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_comments_vakansiya_id ON comments(vakansiya_id);
CREATE INDEX IF NOT EXISTS idx_comments_resume_id    ON comments(resume_id);
CREATE INDEX IF NOT EXISTS idx_comments_parent_id    ON comments(parent_id);
CREATE INDEX IF NOT EXISTS idx_comments_user_id      ON comments(user_id);
CREATE INDEX IF NOT EXISTS idx_comments_deleted_at   ON comments(deleted_at);
