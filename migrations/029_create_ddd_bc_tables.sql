-- +goose Up

-- Feature flags
CREATE TABLE feature_flags (
    id            UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    key           VARCHAR(255) NOT NULL UNIQUE,
    description   TEXT        NOT NULL DEFAULT '',
    flag_type     VARCHAR(20) NOT NULL DEFAULT 'bool',
    default_value TEXT        NOT NULL DEFAULT '',
    active        BOOLEAN     NOT NULL DEFAULT false,
    rollout_pct   INT         NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE feature_flag_rule_groups (
    id         UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    flag_id    UUID        NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
    name       VARCHAR(255) NOT NULL,
    priority   INT         NOT NULL DEFAULT 0,
    value      TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE feature_flag_conditions (
    id             UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    rule_group_id  UUID        NOT NULL REFERENCES feature_flag_rule_groups(id) ON DELETE CASCADE,
    attribute      VARCHAR(255) NOT NULL,
    operator       VARCHAR(20)  NOT NULL,
    value          TEXT         NOT NULL DEFAULT ''
);

-- Site settings
CREATE TABLE site_settings (
    id           UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    key          VARCHAR(255) NOT NULL UNIQUE,
    value        TEXT        NOT NULL DEFAULT '',
    setting_type VARCHAR(50) NOT NULL DEFAULT 'general',
    description  TEXT        NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Data exports
CREATE TABLE data_exports (
    id         UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id    VARCHAR(100) NOT NULL,
    status     VARCHAR(20)  NOT NULL DEFAULT 'PENDING',
    file_url   TEXT        NOT NULL DEFAULT '',
    error_msg  TEXT        NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Statistics snapshots
CREATE TABLE statistics_snapshots (
    id               UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    date             DATE        NOT NULL UNIQUE,
    total_users      BIGINT      NOT NULL DEFAULT 0,
    total_accounts   BIGINT      NOT NULL DEFAULT 0,
    active_accounts  BIGINT      NOT NULL DEFAULT 0,
    total_transfers  BIGINT      NOT NULL DEFAULT 0,
    total_cards      BIGINT      NOT NULL DEFAULT 0,
    pending_kyc      BIGINT      NOT NULL DEFAULT 0,
    flagged_fraud    BIGINT      NOT NULL DEFAULT 0,
    system_errors    BIGINT      NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Notifications
CREATE TABLE notifications (
    id         UUID             NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id    VARCHAR(100)     NOT NULL,
    title      VARCHAR(500)     NOT NULL,
    message    TEXT             NOT NULL DEFAULT '',
    type       VARCHAR(20)      NOT NULL DEFAULT 'INFO',
    data       JSONB            NOT NULL DEFAULT '{}',
    read_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ      NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ      NOT NULL DEFAULT now()
);

-- Translations
CREATE TABLE translations (
    id         UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    key        VARCHAR(500) NOT NULL,
    language   VARCHAR(10)  NOT NULL,
    value      TEXT        NOT NULL DEFAULT '',
    "group"    VARCHAR(100) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(key, language)
);

-- Files
CREATE TABLE files (
    id            UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    name          VARCHAR(500) NOT NULL,
    original_name VARCHAR(500) NOT NULL,
    mime_type     VARCHAR(100) NOT NULL DEFAULT '',
    size          BIGINT       NOT NULL DEFAULT 0,
    path          TEXT         NOT NULL DEFAULT '',
    url           TEXT         NOT NULL DEFAULT '',
    uploaded_by   VARCHAR(100) NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- Announcements
CREATE TABLE announcements (
    id         UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    title_uz   TEXT        NOT NULL DEFAULT '',
    title_ru   TEXT        NOT NULL DEFAULT '',
    title_en   TEXT        NOT NULL DEFAULT '',
    body_uz    TEXT        NOT NULL DEFAULT '',
    body_ru    TEXT        NOT NULL DEFAULT '',
    body_en    TEXT        NOT NULL DEFAULT '',
    priority   INT         NOT NULL DEFAULT 0,
    status     VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    start_date TIMESTAMPTZ,
    end_date   TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- System errors
CREATE TABLE system_errors (
    id          UUID         NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    code        VARCHAR(100) NOT NULL,
    message     TEXT         NOT NULL DEFAULT '',
    severity    VARCHAR(20)  NOT NULL DEFAULT 'MEDIUM',
    category    VARCHAR(20)  NOT NULL DEFAULT 'SYSTEM',
    stack_trace TEXT         NOT NULL DEFAULT '',
    request_id  VARCHAR(100) NOT NULL DEFAULT '',
    user_id     VARCHAR(100) NOT NULL DEFAULT '',
    ip_address  VARCHAR(50)  NOT NULL DEFAULT '',
    path        VARCHAR(500) NOT NULL DEFAULT '',
    method      VARCHAR(10)  NOT NULL DEFAULT '',
    metadata    JSONB        NOT NULL DEFAULT '{}',
    resolution  VARCHAR(20)  NOT NULL DEFAULT 'PENDING',
    resolved_at TIMESTAMPTZ,
    resolved_by VARCHAR(100) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- Error codes
CREATE TABLE error_codes (
    id          UUID         NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    code        VARCHAR(100) NOT NULL UNIQUE,
    message_en  TEXT         NOT NULL DEFAULT '',
    message_uz  TEXT         NOT NULL DEFAULT '',
    message_ru  TEXT         NOT NULL DEFAULT '',
    category    VARCHAR(50)  NOT NULL DEFAULT '',
    severity    VARCHAR(20)  NOT NULL DEFAULT 'MEDIUM',
    http_status INT          NOT NULL DEFAULT 500,
    retryable   BOOLEAN      NOT NULL DEFAULT false,
    suggestion  TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS error_codes;
DROP TABLE IF EXISTS system_errors;
DROP TABLE IF EXISTS announcements;
DROP TABLE IF EXISTS files;
DROP TABLE IF EXISTS translations;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS statistics_snapshots;
DROP TABLE IF EXISTS data_exports;
DROP TABLE IF EXISTS site_settings;
DROP TABLE IF EXISTS feature_flag_conditions;
DROP TABLE IF EXISTS feature_flag_rule_groups;
DROP TABLE IF EXISTS feature_flags;
