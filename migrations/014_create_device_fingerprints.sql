-- +goose Up
CREATE TABLE IF NOT EXISTS device_fingerprints (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_hash VARCHAR(64) NOT NULL,
    device_name TEXT NOT NULL DEFAULT '',
    ip_address  VARCHAR(45) NOT NULL DEFAULT '',
    trusted     BOOLEAN NOT NULL DEFAULT false,
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_device_user_hash UNIQUE (user_id, device_hash)
);

CREATE INDEX idx_device_user ON device_fingerprints(user_id);

-- +goose Down
DROP TABLE IF EXISTS device_fingerprints;
