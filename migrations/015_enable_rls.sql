-- +goose Up

-- Create application role (used by the app, not superuser)
DO $$ BEGIN
    IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'xbank_app') THEN
        CREATE ROLE xbank_app LOGIN;
    END IF;
END $$;

-- Enable RLS on sensitive tables
ALTER TABLE accounts ENABLE ROW LEVEL SECURITY;
ALTER TABLE cards ENABLE ROW LEVEL SECURITY;
ALTER TABLE beneficiaries ENABLE ROW LEVEL SECURITY;
ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;

-- Accounts: user can only see their own accounts
CREATE POLICY accounts_user_policy ON accounts
    FOR ALL
    TO xbank_app
    USING (user_id = current_setting('app.current_user_id', true)::uuid);

-- Cards: user can only see cards on their own accounts
CREATE POLICY cards_user_policy ON cards
    FOR ALL
    TO xbank_app
    USING (account_id IN (
        SELECT id FROM accounts WHERE user_id = current_setting('app.current_user_id', true)::uuid
    ));

-- Beneficiaries: user can only see their own beneficiaries
CREATE POLICY beneficiaries_user_policy ON beneficiaries
    FOR ALL
    TO xbank_app
    USING (user_id = current_setting('app.current_user_id', true)::uuid);

-- Sessions: user can only see their own sessions
CREATE POLICY sessions_user_policy ON sessions
    FOR ALL
    TO xbank_app
    USING (user_id = current_setting('app.current_user_id', true)::uuid);

-- Superuser (postgres) bypasses RLS — for migrations and admin
ALTER TABLE accounts FORCE ROW LEVEL SECURITY;
ALTER TABLE cards FORCE ROW LEVEL SECURITY;
ALTER TABLE beneficiaries FORCE ROW LEVEL SECURITY;
ALTER TABLE sessions FORCE ROW LEVEL SECURITY;

-- +goose Down
DROP POLICY IF EXISTS accounts_user_policy ON accounts;
DROP POLICY IF EXISTS cards_user_policy ON cards;
DROP POLICY IF EXISTS beneficiaries_user_policy ON beneficiaries;
DROP POLICY IF EXISTS sessions_user_policy ON sessions;

ALTER TABLE accounts DISABLE ROW LEVEL SECURITY;
ALTER TABLE cards DISABLE ROW LEVEL SECURITY;
ALTER TABLE beneficiaries DISABLE ROW LEVEL SECURITY;
ALTER TABLE sessions DISABLE ROW LEVEL SECURITY;
