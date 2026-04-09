-- +goose Up

-- Rate Limit Rules (ops/generic/ratelimit)
CREATE TABLE IF NOT EXISTS rate_limit_rules (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key         VARCHAR(255) NOT NULL UNIQUE,
    max_requests   INT NOT NULL,
    window_seconds INT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    enabled     BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- App Metrics (ops/generic/metric)
CREATE TABLE IF NOT EXISTS app_metrics (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         VARCHAR(255) NOT NULL,
    value        DOUBLE PRECISION NOT NULL,
    labels       JSONB NOT NULL DEFAULT '{}',
    collected_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_app_metrics_name ON app_metrics (name);
CREATE INDEX IF NOT EXISTS idx_app_metrics_collected_at ON app_metrics (collected_at DESC);

-- IP Rules (ops/supporting/iprule)
CREATE TABLE IF NOT EXISTS ip_rules (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ip_address VARCHAR(45) NOT NULL,
    rule_type  VARCHAR(10) NOT NULL CHECK (rule_type IN ('ALLOW', 'DENY')),
    reason     TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ,
    created_by VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_ip_rules_ip_address ON ip_rules (ip_address);

-- Integrations (admin/supporting/integration)
CREATE TABLE IF NOT EXISTS integrations (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255) NOT NULL UNIQUE,
    base_url    TEXT NOT NULL,
    api_key     TEXT NOT NULL DEFAULT '',
    status      VARCHAR(20) NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'SUSPENDED')),
    webhook_url TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down

DROP TABLE IF EXISTS integrations;
DROP TABLE IF EXISTS ip_rules;
DROP TABLE IF EXISTS app_metrics;
DROP TABLE IF EXISTS rate_limit_rules;
