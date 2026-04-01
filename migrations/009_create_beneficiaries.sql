-- +goose Up
CREATE TABLE IF NOT EXISTS beneficiaries (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name           VARCHAR(255) NOT NULL,
    account_number VARCHAR(34) NOT NULL,
    bank_name      VARCHAR(255) NOT NULL DEFAULT '',
    bank_code      VARCHAR(11) NOT NULL DEFAULT '',
    currency       VARCHAR(3) NOT NULL DEFAULT 'UZS',
    ben_type       VARCHAR(15) NOT NULL DEFAULT 'INTERNAL',
    verified       BOOLEAN NOT NULL DEFAULT false,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_beneficiary_user_account UNIQUE (user_id, account_number)
);

CREATE INDEX idx_beneficiaries_user_id ON beneficiaries(user_id);

-- +goose Down
DROP TABLE IF EXISTS beneficiaries;
