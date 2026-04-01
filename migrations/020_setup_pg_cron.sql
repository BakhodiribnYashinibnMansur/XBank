-- +goose Up

-- =====================================================================
-- pg_cron — scheduled database jobs.
--
-- Requires: pg_cron extension + shared_preload_libraries = 'pg_cron'
-- See docker-compose.yml for PostgreSQL configuration.
--
-- Jobs:
--   1. Auto-create monthly event partitions (runs 1st of each month)
--   2. Nightly balance snapshot for daily_balance_projections
--   3. Cleanup expired sessions (runs hourly)
-- =====================================================================

CREATE EXTENSION IF NOT EXISTS pg_cron;

-- 1. Auto-create next month's event partitions.
-- Runs at 00:00 on the 1st of every month.
-- Creates partitions 2 months ahead to avoid edge-case races.
SELECT cron.schedule(
    'create-event-partitions',
    '0 0 1 * *',
    $$
    DO $$
    DECLARE
        next_month DATE := date_trunc('month', NOW()) + INTERVAL '1 month';
        two_months DATE := date_trunc('month', NOW()) + INTERVAL '2 months';
        three_months DATE := date_trunc('month', NOW()) + INTERVAL '3 months';
        suffix_1 TEXT := to_char(next_month, 'YYYY"m"MM');
        suffix_2 TEXT := to_char(two_months, 'YYYY"m"MM');
    BEGIN
        -- Account events partitions
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS account_events_y%s PARTITION OF account_events FOR VALUES FROM (%L) TO (%L)',
            suffix_1, next_month, two_months
        );
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS account_events_y%s PARTITION OF account_events FOR VALUES FROM (%L) TO (%L)',
            suffix_2, two_months, three_months
        );

        -- Transfer events partitions
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS transfer_events_y%s PARTITION OF transfer_events FOR VALUES FROM (%L) TO (%L)',
            suffix_1, next_month, two_months
        );
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS transfer_events_y%s PARTITION OF transfer_events FOR VALUES FROM (%L) TO (%L)',
            suffix_2, two_months, three_months
        );
    END
    $$;
    $$
);

-- 2. Nightly balance snapshot — captures EOD balance for every active account.
-- Runs at 23:59 every day.
SELECT cron.schedule(
    'daily-balance-snapshot',
    '59 23 * * *',
    $$
    INSERT INTO daily_balance_projections (account_id, date, balance, currency)
    SELECT id, CURRENT_DATE, balance, currency
    FROM accounts
    WHERE status = 'ACTIVE'
    ON CONFLICT (account_id, date)
    DO UPDATE SET balance = EXCLUDED.balance;
    $$
);

-- 3. Cleanup expired sessions — removes stale sessions from DB.
-- Runs every hour.
SELECT cron.schedule(
    'cleanup-expired-sessions',
    '0 * * * *',
    $$
    DELETE FROM sessions WHERE expires_at < NOW();
    $$
);


-- +goose Down
SELECT cron.unschedule('cleanup-expired-sessions');
SELECT cron.unschedule('daily-balance-snapshot');
SELECT cron.unschedule('create-event-partitions');
DROP EXTENSION IF EXISTS pg_cron;
