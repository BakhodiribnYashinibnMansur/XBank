-- +goose Up
CREATE TABLE IF NOT EXISTS user_settings (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key        VARCHAR(100) NOT NULL,
    value      TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_settings_user_key UNIQUE (user_id, key)
);

CREATE INDEX idx_user_settings_user_id ON user_settings(user_id);

-- +goose Down
DROP TABLE IF EXISTS user_settings;
