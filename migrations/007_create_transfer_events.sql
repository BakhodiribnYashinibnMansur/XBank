-- +goose Up
CREATE TABLE IF NOT EXISTS transfer_events (
    aggregate_id UUID NOT NULL,
    event_type   VARCHAR(30) NOT NULL,
    version      INT NOT NULL,
    attr_key     VARCHAR(50) NOT NULL,
    attr_value   TEXT NOT NULL DEFAULT '',
    occurred_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (aggregate_id, version, attr_key)
);

CREATE INDEX idx_transfer_events_aggregate ON transfer_events(aggregate_id, version);
CREATE INDEX idx_transfer_events_type ON transfer_events(event_type);
CREATE INDEX idx_transfer_events_occurred ON transfer_events(occurred_at);

-- +goose Down
DROP TABLE IF EXISTS transfer_events;
