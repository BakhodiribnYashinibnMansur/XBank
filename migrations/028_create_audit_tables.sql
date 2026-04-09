-- +goose Up

-- Audit logs (partitioned by month)
CREATE TABLE audit_logs (
    id             UUID         NOT NULL DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(50)  NOT NULL,
    aggregate_id   VARCHAR(100) NOT NULL,
    action         VARCHAR(100) NOT NULL,
    actor_id       VARCHAR(100) NOT NULL DEFAULT '',
    attributes     JSONB        NOT NULL DEFAULT '{}',
    ip_address     VARCHAR(45)  NOT NULL DEFAULT '',
    user_agent     TEXT         NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE TABLE audit_logs_default PARTITION OF audit_logs DEFAULT;
CREATE TABLE audit_logs_y2026m04 PARTITION OF audit_logs FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE audit_logs_y2026m05 PARTITION OF audit_logs FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE audit_logs_y2026m06 PARTITION OF audit_logs FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE audit_logs_y2026m07 PARTITION OF audit_logs FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');

CREATE INDEX idx_audit_logs_aggregate ON audit_logs (aggregate_type, aggregate_id);
CREATE INDEX idx_audit_logs_actor     ON audit_logs (actor_id);
CREATE INDEX idx_audit_logs_action    ON audit_logs (action);
CREATE INDEX idx_audit_logs_created   ON audit_logs (created_at);

-- Endpoint history (partitioned by month)
CREATE TABLE endpoint_history (
    id          UUID         NOT NULL DEFAULT gen_random_uuid(),
    method      VARCHAR(10)  NOT NULL,
    path        VARCHAR(500) NOT NULL,
    status_code INT          NOT NULL,
    user_id     VARCHAR(100) NOT NULL DEFAULT '',
    ip_address  VARCHAR(45)  NOT NULL DEFAULT '',
    duration_ms INT          NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE TABLE endpoint_history_default PARTITION OF endpoint_history DEFAULT;
CREATE TABLE endpoint_history_y2026m04 PARTITION OF endpoint_history FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE endpoint_history_y2026m05 PARTITION OF endpoint_history FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE endpoint_history_y2026m06 PARTITION OF endpoint_history FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE endpoint_history_y2026m07 PARTITION OF endpoint_history FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');

CREATE INDEX idx_endpoint_history_user    ON endpoint_history (user_id);
CREATE INDEX idx_endpoint_history_path    ON endpoint_history (path, method);
CREATE INDEX idx_endpoint_history_created ON endpoint_history (created_at);

-- +goose Down
DROP TABLE IF EXISTS endpoint_history CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
