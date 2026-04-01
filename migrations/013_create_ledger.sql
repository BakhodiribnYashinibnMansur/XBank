-- +goose Up
CREATE TABLE IF NOT EXISTS ledger_entries (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id  UUID NOT NULL,
    transfer_id UUID NOT NULL,
    entry_type  VARCHAR(6) NOT NULL, -- DEBIT or CREDIT
    amount      BIGINT NOT NULL,
    currency    VARCHAR(3) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_ledger_positive CHECK (amount > 0)
);

CREATE INDEX idx_ledger_account ON ledger_entries(account_id, created_at DESC);
CREATE INDEX idx_ledger_transfer ON ledger_entries(transfer_id);

-- +goose Down
DROP TABLE IF EXISTS ledger_entries;
