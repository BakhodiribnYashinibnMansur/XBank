-- +goose Up
CREATE TABLE IF NOT EXISTS cards (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id    UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    pan           VARCHAR(16) UNIQUE NOT NULL,
    masked_pan    VARCHAR(25) NOT NULL,
    pin_hash      VARCHAR(255) NOT NULL DEFAULT '',
    expiry_month  SMALLINT NOT NULL,
    expiry_year   SMALLINT NOT NULL,
    card_type     VARCHAR(10) NOT NULL DEFAULT 'DEBIT',
    status        VARCHAR(12) NOT NULL DEFAULT 'INACTIVE',
    pin_attempts  SMALLINT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cards_account_id ON cards(account_id);
CREATE INDEX idx_cards_pan ON cards(pan);

-- +goose Down
DROP INDEX IF EXISTS idx_cards_pan;
DROP INDEX IF EXISTS idx_cards_account_id;
DROP TABLE IF EXISTS cards;
