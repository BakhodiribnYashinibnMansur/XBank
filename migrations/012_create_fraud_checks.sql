-- +goose Up
CREATE TABLE IF NOT EXISTS fraud_checks (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transfer_id    UUID NOT NULL,
    user_id        UUID NOT NULL,
    risk_score     INT NOT NULL DEFAULT 0,
    risk_level     VARCHAR(10) NOT NULL DEFAULT 'LOW',
    action         VARCHAR(10) NOT NULL DEFAULT 'APPROVE',
    flags          TEXT[] NOT NULL DEFAULT '{}',
    reviewed_by    UUID,
    review_comment TEXT NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fraud_transfer ON fraud_checks(transfer_id);
CREATE INDEX idx_fraud_action ON fraud_checks(action);

-- +goose Down
DROP TABLE IF EXISTS fraud_checks;
