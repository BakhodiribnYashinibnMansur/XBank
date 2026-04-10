-- +goose Up

-- Currencies (banking/generic/currency)
CREATE TABLE IF NOT EXISTS currencies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code            VARCHAR(3)   NOT NULL UNIQUE,
    name            VARCHAR(100) NOT NULL,
    symbol          VARCHAR(10)  NOT NULL DEFAULT '',
    decimal_places  INT          NOT NULL DEFAULT 2,
    status          VARCHAR(20)  NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE')),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_currencies_code ON currencies (code);

-- Templates (content/core/template)
CREATE TABLE IF NOT EXISTS templates (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug        VARCHAR(255) NOT NULL,
    channel     VARCHAR(10)  NOT NULL CHECK (channel IN ('EMAIL', 'SMS', 'PUSH')),
    subject     TEXT         NOT NULL DEFAULT '',
    body        TEXT         NOT NULL,
    locale      VARCHAR(10)  NOT NULL DEFAULT 'en',
    status      VARCHAR(20)  NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'ACTIVE', 'ARCHIVED')),
    version     INT          NOT NULL DEFAULT 1,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_templates_slug_locale ON templates (slug, locale);

-- Health Records (ops/core/healthcheck)
CREATE TABLE IF NOT EXISTS health_records (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status      VARCHAR(20)  NOT NULL CHECK (status IN ('HEALTHY', 'DEGRADED', 'UNHEALTHY')),
    components  JSONB        NOT NULL DEFAULT '[]',
    checked_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_health_records_checked_at ON health_records (checked_at DESC);

-- +goose Down
DROP TABLE IF EXISTS health_records;
DROP TABLE IF EXISTS templates;
DROP TABLE IF EXISTS currencies;
