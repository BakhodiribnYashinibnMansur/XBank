-- +goose Up
CREATE TABLE IF NOT EXISTS exchange_rates (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_currency VARCHAR(3) NOT NULL,
    to_currency   VARCHAR(3) NOT NULL,
    buy_rate      BIGINT NOT NULL,
    sell_rate     BIGINT NOT NULL,
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_exchange_pair UNIQUE (from_currency, to_currency)
);

-- +goose Down
DROP TABLE IF EXISTS exchange_rates;
