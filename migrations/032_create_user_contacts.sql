-- +goose Up
CREATE TABLE IF NOT EXISTS user_contacts (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id   UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    contact_id UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    custom_name VARCHAR(100) NOT NULL DEFAULT '',
    is_blocked BOOLEAN     NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_user_contacts_owner_contact UNIQUE (owner_id, contact_id)
);

CREATE INDEX idx_user_contacts_owner ON user_contacts(owner_id);

-- +goose Down
DROP TABLE IF EXISTS user_contacts;
