-- +goose Up

-- TOTP (Time-based One-Time Password) fields for 2FA.
-- totp_secret: base32-encoded shared secret (encrypted at rest via app layer)
-- totp_enabled: whether 2FA is active for this user
-- totp_verified_at: when the user first verified their TOTP setup
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_secret      VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_enabled     BOOLEAN      NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS totp_verified_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS totp_verified_at;
ALTER TABLE users DROP COLUMN IF EXISTS totp_enabled;
ALTER TABLE users DROP COLUMN IF EXISTS totp_secret;
