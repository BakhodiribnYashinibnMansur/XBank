-- +goose Up
CREATE TABLE IF NOT EXISTS kyc_verifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID UNIQUE NOT NULL REFERENCES users(id),
    document_type   VARCHAR(20) NOT NULL,
    document_number VARCHAR(255) NOT NULL,
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL DEFAULT '',
    date_of_birth   VARCHAR(10) NOT NULL DEFAULT '',
    status          VARCHAR(10) NOT NULL DEFAULT 'PENDING',
    rejected_reason TEXT NOT NULL DEFAULT '',
    reviewed_by     UUID,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_kyc_status ON kyc_verifications(status);

-- +goose Down
DROP TABLE IF EXISTS kyc_verifications;
