-- +goose Up
-- PAN column must accommodate encrypted values (AES-GCM output > 16 chars)
ALTER TABLE cards ALTER COLUMN pan TYPE VARCHAR(255);

-- +goose Down
ALTER TABLE cards ALTER COLUMN pan TYPE VARCHAR(16);
