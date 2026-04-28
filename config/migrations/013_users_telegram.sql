ALTER TABLE users
    ALTER COLUMN phone    DROP NOT NULL,
    ALTER COLUMN password DROP NOT NULL;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS telegram_id       BIGINT UNIQUE,
    ADD COLUMN IF NOT EXISTS telegram_username VARCHAR(255);

CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);

CREATE TABLE IF NOT EXISTS tg_login_sessions (
    token        VARCHAR(64)  PRIMARY KEY,
    telegram_id  BIGINT,
    user_id      BIGINT       REFERENCES users(id) ON DELETE SET NULL,
    status       VARCHAR(20)  NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ  NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tg_sessions_status  ON tg_login_sessions(status);
CREATE INDEX IF NOT EXISTS idx_tg_sessions_expires ON tg_login_sessions(expires_at);
