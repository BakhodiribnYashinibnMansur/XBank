-- +goose Up

-- Card tokenization — maps opaque tokens to encrypted PANs.
-- PAN never leaves the vault; merchants and APIs use tokens instead.
-- Token format: tok_xxxxxxxxxxxxxxxx (24 chars, prefix + 20 hex)
CREATE TABLE IF NOT EXISTS card_tokens (
    token       VARCHAR(24) PRIMARY KEY,
    card_id     UUID        NOT NULL REFERENCES cards(id) ON DELETE CASCADE,
    pan_encrypted TEXT      NOT NULL,  -- AES-256-GCM encrypted PAN
    last_four   VARCHAR(4)  NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_active   BOOLEAN     NOT NULL DEFAULT TRUE
);

CREATE INDEX idx_card_tokens_card ON card_tokens(card_id);
CREATE INDEX idx_card_tokens_active ON card_tokens(is_active, expires_at) WHERE is_active = TRUE;

-- Add 3DS fields to cards table
ALTER TABLE cards ADD COLUMN IF NOT EXISTS three_ds_enrolled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE cards ADD COLUMN IF NOT EXISTS three_ds_version  VARCHAR(5) NOT NULL DEFAULT '';
ALTER TABLE cards ADD COLUMN IF NOT EXISTS emv_aid           VARCHAR(32) NOT NULL DEFAULT '';

-- Hold/capture/release: card authorization holds
CREATE TABLE IF NOT EXISTS card_holds (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id      UUID        NOT NULL REFERENCES cards(id),
    account_id   UUID        NOT NULL REFERENCES accounts(id),
    merchant     VARCHAR(100) NOT NULL DEFAULT '',
    amount       BIGINT      NOT NULL,
    currency     VARCHAR(3)  NOT NULL,
    status       VARCHAR(10) NOT NULL DEFAULT 'HELD',  -- HELD, CAPTURED, RELEASED, EXPIRED
    held_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMPTZ NOT NULL,
    captured_at  TIMESTAMPTZ,
    released_at  TIMESTAMPTZ,
    CONSTRAINT chk_hold_positive CHECK (amount > 0)
);

CREATE INDEX idx_card_holds_card   ON card_holds(card_id, status);
CREATE INDEX idx_card_holds_acct   ON card_holds(account_id, status);
CREATE INDEX idx_card_holds_expire ON card_holds(status, expires_at) WHERE status = 'HELD';

-- +goose Down
ALTER TABLE cards DROP COLUMN IF EXISTS emv_aid;
ALTER TABLE cards DROP COLUMN IF EXISTS three_ds_version;
ALTER TABLE cards DROP COLUMN IF EXISTS three_ds_enrolled;
DROP TABLE IF EXISTS card_holds;
DROP TABLE IF EXISTS card_tokens;
