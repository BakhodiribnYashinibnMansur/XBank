-- +goose Up

-- =====================================================================
-- Event store partitioning by occurred_at (monthly range partitions).
--
-- Why partition?
--   1. Event tables grow unbounded (every deposit/withdrawal = new rows).
--   2. Range scans on occurred_at become O(partition) instead of O(table).
--   3. Old partitions can be archived/detached without downtime.
--   4. pg_cron auto-creates future partitions (see 020_setup_pg_cron.sql).
--
-- Strategy: rename old table → create partitioned parent → copy data →
--           create initial monthly partitions.
-- =====================================================================

-- === Account Events ===

-- 1. Rename the original table
ALTER TABLE account_events RENAME TO account_events_old;

-- 2. Create partitioned parent (same schema, partitioned by RANGE on occurred_at)
CREATE TABLE account_events (
    aggregate_id UUID        NOT NULL,
    event_type   VARCHAR(30) NOT NULL,
    version      INT         NOT NULL,
    attr_key     VARCHAR(50) NOT NULL,
    attr_value   TEXT        NOT NULL DEFAULT '',
    occurred_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (aggregate_id, version, attr_key, occurred_at)
) PARTITION BY RANGE (occurred_at);

-- 3. Default partition (catches anything outside defined ranges)
CREATE TABLE account_events_default PARTITION OF account_events DEFAULT;

-- 4. Initial monthly partitions (current + next 3 months)
CREATE TABLE account_events_y2026m04 PARTITION OF account_events
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE account_events_y2026m05 PARTITION OF account_events
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE account_events_y2026m06 PARTITION OF account_events
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE account_events_y2026m07 PARTITION OF account_events
    FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');

-- 5. Copy existing data into partitioned table
INSERT INTO account_events SELECT * FROM account_events_old;

-- 6. Drop old table
DROP TABLE account_events_old;

-- 7. Recreate indexes on partitioned table
CREATE INDEX idx_acce_aggregate ON account_events(aggregate_id, version);
CREATE INDEX idx_acce_type      ON account_events(event_type);
CREATE INDEX idx_acce_occurred  ON account_events(occurred_at);

-- === Transfer Events ===

ALTER TABLE transfer_events RENAME TO transfer_events_old;

CREATE TABLE transfer_events (
    aggregate_id UUID        NOT NULL,
    event_type   VARCHAR(30) NOT NULL,
    version      INT         NOT NULL,
    attr_key     VARCHAR(50) NOT NULL,
    attr_value   TEXT        NOT NULL DEFAULT '',
    occurred_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (aggregate_id, version, attr_key, occurred_at)
) PARTITION BY RANGE (occurred_at);

CREATE TABLE transfer_events_default PARTITION OF transfer_events DEFAULT;

CREATE TABLE transfer_events_y2026m04 PARTITION OF transfer_events
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');
CREATE TABLE transfer_events_y2026m05 PARTITION OF transfer_events
    FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE transfer_events_y2026m06 PARTITION OF transfer_events
    FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE transfer_events_y2026m07 PARTITION OF transfer_events
    FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');

INSERT INTO transfer_events SELECT * FROM transfer_events_old;

DROP TABLE transfer_events_old;

CREATE INDEX idx_trfe_aggregate ON transfer_events(aggregate_id, version);
CREATE INDEX idx_trfe_type      ON transfer_events(event_type);
CREATE INDEX idx_trfe_occurred  ON transfer_events(occurred_at);


-- +goose Down

-- Reverse: create non-partitioned tables, copy data back

-- Account events
ALTER TABLE account_events RENAME TO account_events_partitioned;

CREATE TABLE account_events (
    aggregate_id UUID        NOT NULL,
    event_type   VARCHAR(30) NOT NULL,
    version      INT         NOT NULL,
    attr_key     VARCHAR(50) NOT NULL,
    attr_value   TEXT        NOT NULL DEFAULT '',
    occurred_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (aggregate_id, version, attr_key)
);

INSERT INTO account_events SELECT aggregate_id, event_type, version, attr_key, attr_value, occurred_at
FROM account_events_partitioned;

DROP TABLE account_events_partitioned CASCADE;

CREATE INDEX idx_account_events_aggregate ON account_events(aggregate_id, version);
CREATE INDEX idx_account_events_type      ON account_events(event_type);
CREATE INDEX idx_account_events_occurred  ON account_events(occurred_at);
CREATE INDEX idx_account_events_attr      ON account_events(attr_key, attr_value);

-- Transfer events
ALTER TABLE transfer_events RENAME TO transfer_events_partitioned;

CREATE TABLE transfer_events (
    aggregate_id UUID        NOT NULL,
    event_type   VARCHAR(30) NOT NULL,
    version      INT         NOT NULL,
    attr_key     VARCHAR(50) NOT NULL,
    attr_value   TEXT        NOT NULL DEFAULT '',
    occurred_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (aggregate_id, version, attr_key)
);

INSERT INTO transfer_events SELECT aggregate_id, event_type, version, attr_key, attr_value, occurred_at
FROM transfer_events_partitioned;

DROP TABLE transfer_events_partitioned CASCADE;

CREATE INDEX idx_transfer_events_aggregate ON transfer_events(aggregate_id, version);
CREATE INDEX idx_transfer_events_type      ON transfer_events(event_type);
CREATE INDEX idx_transfer_events_occurred  ON transfer_events(occurred_at);
