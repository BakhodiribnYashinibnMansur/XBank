# PostgreSQL — Bank Database Standard

## Schema Design Rules
```
- UUID — primary key (not auto-increment, not guessable)
- created_at, updated_at — in every table (TIMESTAMPTZ)
- Soft delete — deleted_at (financial data is NEVER hard deleted)
- BIGINT — for money (minor units: tiyin/cent), NEVER FLOAT!
- CHECK constraints — balance_minor >= 0, amount > 0
- UNIQUE constraints — idempotency_key, account_number, email
- FK constraints — referential integrity
```

## Row Level Security (RLS)
```sql
ALTER TABLE accounts ENABLE ROW LEVEL SECURITY;
CREATE POLICY accounts_owner_policy ON accounts
    FOR ALL USING (user_id = current_setting('app.current_user_id')::uuid);

-- Before each query:
SET LOCAL app.current_user_id = 'user-uuid';
-- ADMIN: BYPASSRLS
```

Applied to tables: accounts, cards, beneficiaries, notifications

## Encryption at Rest

```sql
CREATE EXTENSION IF NOT EXISTS pgcrypto;
-- Application-level Hybrid/Envelope encryption (key in Vault)
```

### Encryption Keys Table
```sql
CREATE TABLE encryption_keys (
    id              VARCHAR(50) PRIMARY KEY,  -- 'jwt_2026_q1', 'card_kek_v3'
    purpose         VARCHAR(30) NOT NULL,     -- JWT, CARD_KEK, KYC_KEK
    algorithm       VARCHAR(20) NOT NULL,     -- ES256, RSA-4096
    public_key_pem  TEXT NOT NULL,            -- PEM format
    -- private_key is NEVER STORED IN DB! → Vault / HSM
    key_version     INTEGER NOT NULL DEFAULT 1,
    status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
        -- ACTIVE, ROTATE_OUT, RETIRED
    activated_at    TIMESTAMPTZ NOT NULL,
    rotate_after    TIMESTAMPTZ NOT NULL,
    retired_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_enc_keys_purpose ON encryption_keys (purpose, status);
```

### User Signing Keys Table (ECDSA per-user)
```sql
CREATE TABLE user_signing_keys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    device_id       VARCHAR(255) NOT NULL,
    algorithm       VARCHAR(20) NOT NULL DEFAULT 'ES256',
    public_key_pem  TEXT NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
        -- ACTIVE, REVOKED
    activated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL,
    revoked_at      TIMESTAMPTZ,
    last_used_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, device_id, status)
);

CREATE INDEX idx_user_sign_keys ON user_signing_keys (user_id, status)
    WHERE status = 'ACTIVE';
```

Details: [Encryption & PKI](../security/encryption.md)

## SSL Connection
```
DB_SSLMODE=verify-full
-- pg_hba.conf:
hostssl all all 0.0.0.0/0 scram-sha-256   # SSL only
hostnossl all all 0.0.0.0/0 reject         # reject without SSL
```

## Least Privilege — Separate DB Users
```sql
-- xbank_app: SELECT, INSERT, UPDATE (no DELETE!)
-- xbank_readonly: SELECT only (CQRS read side)
-- xbank_migrate: ALL (only during deploy)
-- audit_log, event_store: UPDATE/DELETE REVOKE
```

## PgBouncer — Connection Pooling
```
pool_mode = transaction (compatible with DDD UoW)
max_client_conn = 1000
default_pool_size = 25
```

## Indexes
```sql
-- Standard
CREATE INDEX idx_accounts_user ON accounts(user_id);
CREATE INDEX idx_transfers_from ON transfers(from_account_id, created_at DESC);

-- Partial (smaller, faster)
CREATE INDEX idx_active_accounts ON accounts(user_id)
    WHERE status = 'ACTIVE' AND deleted_at IS NULL;
CREATE INDEX idx_pending_transfers ON transfers(status, created_at)
    WHERE status IN ('PENDING', 'PROCESSING');

-- Covering (index-only scan)
CREATE INDEX idx_account_balance ON accounts(id)
    INCLUDE (balance_minor, available_balance, hold_amount, version);
```

## Table Partitioning (monthly)
- event_store, ledger_entries, audit_log, transfers
- Automatic new partition creation with pg_cron

## Read Replicas (CQRS)
```go
type DB struct {
    WritePool *pgxpool.Pool  // primary
    ReadPool  *pgxpool.Pool  // replica
}
```

## Backup & Disaster Recovery
- WAL archiving → PITR
- pg_basebackup → daily
- Streaming replication
- RPO < 1 minute, RTO < 15 minutes

## pg_cron — Scheduled Jobs
- Clean expired sessions (every hour)
- Delete expired OTPs (every 5 min)
- Clean idempotency keys (daily)
- Create new partitions (monthly)
- Reconciliation (daily at 3:00)

## pg_stat_statements — Query Monitoring
```sql
SELECT query, calls, mean_exec_time FROM pg_stat_statements
ORDER BY mean_exec_time DESC LIMIT 10;
```
