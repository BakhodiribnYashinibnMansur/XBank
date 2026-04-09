-- +goose Up
CREATE TABLE outbox (
    id             BIGSERIAL    PRIMARY KEY,
    aggregate_type VARCHAR(50)  NOT NULL,
    aggregate_id   VARCHAR(100) NOT NULL,
    topic          VARCHAR(100) NOT NULL,
    partition_key  VARCHAR(100) NOT NULL,
    payload        BYTEA        NOT NULL,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_outbox_created_at ON outbox (created_at ASC);

-- +goose Down
DROP TABLE IF EXISTS outbox;
