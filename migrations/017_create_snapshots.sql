-- +goose Up

-- Account snapshots — periodic materialized state checkpoints.
-- Instead of replaying all events, load latest snapshot + events after it.
-- Stored separately from events for clean separation and faster lookups.
CREATE TABLE IF NOT EXISTS account_snapshots (
    aggregate_id UUID        NOT NULL PRIMARY KEY,
    version      INT         NOT NULL,
    user_id      UUID        NOT NULL,
    account_number VARCHAR(20) NOT NULL,
    balance      BIGINT      NOT NULL DEFAULT 0,
    currency     VARCHAR(3)  NOT NULL,
    status       VARCHAR(10) NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_account_snapshots_user ON account_snapshots(user_id);

-- Transfer snapshots — same pattern for transfer aggregates.
CREATE TABLE IF NOT EXISTS transfer_snapshots (
    aggregate_id    UUID        NOT NULL PRIMARY KEY,
    version         INT         NOT NULL,
    from_account_id UUID        NOT NULL,
    to_account_id   UUID        NOT NULL,
    amount          BIGINT      NOT NULL,
    currency        VARCHAR(3)  NOT NULL,
    status          VARCHAR(10) NOT NULL,
    description     TEXT        NOT NULL DEFAULT '',
    failure_reason  TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transfer_snapshots_from ON transfer_snapshots(from_account_id);
CREATE INDEX idx_transfer_snapshots_to   ON transfer_snapshots(to_account_id);

-- Clean up old Snapshot rows from account_events (they were stored inline before).
-- This is safe because the new table replaces them.
DELETE FROM account_events WHERE event_type = 'Snapshot';

-- +goose Down
DROP TABLE IF EXISTS transfer_snapshots;
DROP TABLE IF EXISTS account_snapshots;
