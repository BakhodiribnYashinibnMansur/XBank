-- +goose Up

-- Scheduled transfers — transfers that execute at a future date/time.
-- A background worker picks them up when execute_at <= NOW().
CREATE TABLE IF NOT EXISTS scheduled_transfers (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    from_account_id UUID        NOT NULL REFERENCES accounts(id),
    to_account_id   UUID        NOT NULL REFERENCES accounts(id),
    amount          BIGINT      NOT NULL,
    currency        VARCHAR(3)  NOT NULL,
    description     TEXT        NOT NULL DEFAULT '',
    status          VARCHAR(10) NOT NULL DEFAULT 'PENDING',
    execute_at      TIMESTAMPTZ NOT NULL,
    transfer_id     UUID,                           -- filled after execution
    failure_reason  TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    executed_at     TIMESTAMPTZ,
    CONSTRAINT chk_sched_positive CHECK (amount > 0),
    CONSTRAINT chk_sched_different CHECK (from_account_id != to_account_id),
    CONSTRAINT chk_sched_status CHECK (status IN ('PENDING','EXECUTED','FAILED','CANCELLED'))
);

CREATE INDEX idx_sched_due    ON scheduled_transfers(status, execute_at) WHERE status = 'PENDING';
CREATE INDEX idx_sched_user   ON scheduled_transfers(user_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS scheduled_transfers;
