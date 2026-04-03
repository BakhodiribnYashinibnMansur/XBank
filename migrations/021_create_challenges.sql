-- +goose Up

-- Step-up authentication challenges.
-- Used for sensitive operations (large transfers, card issuance, etc.)
-- that require additional identity verification beyond JWT.
CREATE TABLE IF NOT EXISTS challenges (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    method      VARCHAR(20) NOT NULL DEFAULT 'PASSWORD',
    status      VARCHAR(10) NOT NULL DEFAULT 'PENDING',
    token       VARCHAR(64) UNIQUE,
    action      VARCHAR(50) NOT NULL,
    metadata    TEXT        NOT NULL DEFAULT '',
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMPTZ
);

CREATE INDEX idx_challenges_user_id   ON challenges(user_id, status);
CREATE INDEX idx_challenges_token     ON challenges(token) WHERE token IS NOT NULL;
CREATE INDEX idx_challenges_expires   ON challenges(expires_at);

-- +goose Down
DROP TABLE IF EXISTS challenges;
