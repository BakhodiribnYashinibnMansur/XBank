-- +goose Up
CREATE TABLE IF NOT EXISTS admin_whitelist_ips (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ip_address VARCHAR(45) NOT NULL UNIQUE,
    label      VARCHAR(100) NOT NULL DEFAULT '',
    created_by UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS admin_whitelist_ips;
