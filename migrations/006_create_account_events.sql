-- +goose Up
CREATE TABLE IF NOT EXISTS account_events (
    aggregate_id UUID NOT NULL,
    event_type   VARCHAR(30) NOT NULL,
    version      INT NOT NULL,
    attr_key     VARCHAR(50) NOT NULL,
    attr_value   TEXT NOT NULL DEFAULT '',
    occurred_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (aggregate_id, version, attr_key)
);

CREATE INDEX idx_account_events_aggregate ON account_events(aggregate_id, version);
CREATE INDEX idx_account_events_type ON account_events(event_type);
CREATE INDEX idx_account_events_occurred ON account_events(occurred_at);
CREATE INDEX idx_account_events_attr ON account_events(attr_key, attr_value);

-- +goose Down
DROP TABLE IF EXISTS account_events;
