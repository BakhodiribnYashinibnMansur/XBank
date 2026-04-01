-- +goose Up

-- =====================================================================
-- CQRS Read Projections — denormalized, read-optimized views.
--
-- Why separate read models?
--   1. Command side (accounts, transfers) uses SERIALIZABLE isolation.
--   2. Read side can use read replicas with eventual consistency.
--   3. Denormalized structure avoids JOINs for common queries.
--   4. Different indexes optimized for dashboard/list/search use cases.
--
-- These projections are rebuilt from events and can be recreated at
-- any time without data loss (event store is the source of truth).
-- =====================================================================

-- Account summary projection — dashboard view per user.
-- Denormalized: includes user email so no JOIN needed for display.
CREATE TABLE IF NOT EXISTS account_projections (
    id              UUID        PRIMARY KEY,
    user_id         UUID        NOT NULL,
    user_email      VARCHAR(255) NOT NULL DEFAULT '',
    account_number  VARCHAR(20) NOT NULL,
    balance         BIGINT      NOT NULL DEFAULT 0,
    currency        VARCHAR(3)  NOT NULL,
    status          VARCHAR(10) NOT NULL DEFAULT 'ACTIVE',
    total_credited  BIGINT      NOT NULL DEFAULT 0,  -- lifetime deposits
    total_debited   BIGINT      NOT NULL DEFAULT 0,  -- lifetime withdrawals
    event_count     INT         NOT NULL DEFAULT 0,   -- total events applied
    last_event_at   TIMESTAMPTZ,                      -- last activity time
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_acc_proj_user      ON account_projections(user_id);
CREATE INDEX idx_acc_proj_status    ON account_projections(status);
CREATE INDEX idx_acc_proj_currency  ON account_projections(currency);
CREATE INDEX idx_acc_proj_updated   ON account_projections(updated_at DESC);

-- Transfer history projection — optimized for user-facing transfer list.
-- Denormalized: includes account numbers so no JOIN needed.
CREATE TABLE IF NOT EXISTS transfer_projections (
    id                  UUID        PRIMARY KEY,
    from_account_id     UUID        NOT NULL,
    from_account_number VARCHAR(20) NOT NULL DEFAULT '',
    from_user_id        UUID        NOT NULL,
    to_account_id       UUID        NOT NULL,
    to_account_number   VARCHAR(20) NOT NULL DEFAULT '',
    to_user_id          UUID        NOT NULL,
    amount              BIGINT      NOT NULL,
    currency            VARCHAR(3)  NOT NULL,
    status              VARCHAR(10) NOT NULL DEFAULT 'PENDING',
    description         TEXT        NOT NULL DEFAULT '',
    failure_reason      TEXT        NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at        TIMESTAMPTZ
);

CREATE INDEX idx_trf_proj_from_user ON transfer_projections(from_user_id, created_at DESC);
CREATE INDEX idx_trf_proj_to_user   ON transfer_projections(to_user_id, created_at DESC);
CREATE INDEX idx_trf_proj_from_acc  ON transfer_projections(from_account_id, created_at DESC);
CREATE INDEX idx_trf_proj_to_acc    ON transfer_projections(to_account_id, created_at DESC);
CREATE INDEX idx_trf_proj_status    ON transfer_projections(status, created_at DESC);

-- Daily account balance history — for charts and reporting.
-- One row per account per day, updated by pg_cron nightly job.
CREATE TABLE IF NOT EXISTS daily_balance_projections (
    account_id UUID        NOT NULL,
    date       DATE        NOT NULL,
    balance    BIGINT      NOT NULL,
    currency   VARCHAR(3)  NOT NULL,
    PRIMARY KEY (account_id, date)
);

CREATE INDEX idx_daily_bal_date ON daily_balance_projections(date DESC);

-- Seed initial projection data from existing accounts and transfers
INSERT INTO account_projections (id, user_id, account_number, balance, currency, status, created_at, updated_at)
SELECT id, user_id, account_number, balance, currency, status, created_at, updated_at
FROM accounts
ON CONFLICT (id) DO NOTHING;

INSERT INTO transfer_projections (id, from_account_id, from_user_id, to_account_id, to_user_id, amount, currency, status, description, failure_reason, created_at)
SELECT
    t.id,
    t.from_account_id,
    fa.user_id,
    t.to_account_id,
    ta.user_id,
    t.amount,
    t.currency,
    t.status,
    t.description,
    t.failure_reason,
    t.created_at
FROM transfers t
JOIN accounts fa ON fa.id = t.from_account_id
JOIN accounts ta ON ta.id = t.to_account_id
ON CONFLICT (id) DO NOTHING;


-- +goose Down
DROP TABLE IF EXISTS daily_balance_projections;
DROP TABLE IF EXISTS transfer_projections;
DROP TABLE IF EXISTS account_projections;
