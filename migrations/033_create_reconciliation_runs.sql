-- +goose Up
CREATE TABLE IF NOT EXISTS reconciliation_runs (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID        NOT NULL REFERENCES users(id),
    total_checked  INT         NOT NULL DEFAULT 0,
    matches        INT         NOT NULL DEFAULT 0,
    mismatches     INT         NOT NULL DEFAULT 0,
    status         VARCHAR(20) NOT NULL DEFAULT 'COMPLETED',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reconciliation_runs_user ON reconciliation_runs(user_id);

-- +goose Down
DROP TABLE IF EXISTS reconciliation_runs;
