-- +goose Up

-- Dead Letter Queue — stores failed Kafka messages for retry.
-- When a Kafka publish fails, the message lands here instead of being lost.
-- A pg_cron job retries pending messages periodically.
CREATE TABLE IF NOT EXISTS dead_letter_queue (
    id           BIGSERIAL   PRIMARY KEY,
    topic        VARCHAR(100) NOT NULL,
    partition_key VARCHAR(100) NOT NULL,
    payload      BYTEA       NOT NULL,
    error_msg    TEXT        NOT NULL DEFAULT '',
    retry_count  INT         NOT NULL DEFAULT 0,
    max_retries  INT         NOT NULL DEFAULT 5,
    status       VARCHAR(10) NOT NULL DEFAULT 'PENDING',  -- PENDING, DELIVERED, DEAD
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    next_retry   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMPTZ
);

CREATE INDEX idx_dlq_status_retry ON dead_letter_queue(status, next_retry) WHERE status = 'PENDING';
CREATE INDEX idx_dlq_topic        ON dead_letter_queue(topic);
CREATE INDEX idx_dlq_created      ON dead_letter_queue(created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS dead_letter_queue;
