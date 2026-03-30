# PostgreSQL Ishlash Flow

## Connection Flow

```
Go App
  │
  ├── Write queries ──→ PgBouncer (:6432) ──→ PostgreSQL Primary (:5432)
  │                     (transaction mode)
  │
  └── Read queries ───→ PgBouncer (:6432) ──→ PostgreSQL Replica (:5433)
                        (CQRS query side)
```

## Transaction Flow (Transfer)

```sql
-- 1. Begin SERIALIZABLE
BEGIN ISOLATION LEVEL SERIALIZABLE;

-- 2. Set RLS context
SET LOCAL app.current_user_id = 'user-uuid';

-- 3. Lock accounts (UUID tartibida — deadlock prevention)
SELECT * FROM accounts WHERE id = 'smaller-uuid' FOR UPDATE;
SELECT * FROM accounts WHERE id = 'larger-uuid' FOR UPDATE;

-- 4. Balance check
-- (SERIALIZABLE = write skew himoya)

-- 5. Debit source (optimistic lock)
UPDATE accounts
SET balance_minor = balance_minor - 100000,
    available_balance = available_balance - 100000,
    version = version + 1,
    updated_at = NOW()
WHERE id = 'source' AND version = 5;
-- 0 rows → retry!

-- 6. Credit target
UPDATE accounts
SET balance_minor = balance_minor + 100000,
    available_balance = available_balance + 100000,
    version = version + 1,
    updated_at = NOW()
WHERE id = 'target' AND version = 3;

-- 7. Event store (append-only)
INSERT INTO event_store (aggregate_id, event_type, event_data, version)
VALUES ('source', 'AccountDebited', '{"amount":100000}', 6);
INSERT INTO event_store (aggregate_id, event_type, event_data, version)
VALUES ('target', 'AccountCredited', '{"amount":100000}', 4);

-- 8. Ledger entries (double-entry)
INSERT INTO ledger_entries (transfer_id, account_id, side, amount_minor)
VALUES ('txn-id', 'source', 'DEBIT', 100000);
INSERT INTO ledger_entries (transfer_id, account_id, side, amount_minor)
VALUES ('txn-id', 'target', 'CREDIT', 100000);

-- 9. Transfer record
INSERT INTO transfers (id, from_account_id, to_account_id, amount_minor, status)
VALUES ('txn-id', 'source', 'target', 100000, 'COMPLETED');

-- 10. Audit log
INSERT INTO audit_log (actor_id, action, resource_type, resource_id, new_value)
VALUES ('user-id', 'TRANSFER_COMPLETED', 'TRANSFER', 'txn-id', '{"amount":100000}');

-- 11. Idempotency key
INSERT INTO idempotency_keys (key, user_id, response_code, response_body)
VALUES ('idemp-key', 'user-id', 201, '{"transfer_id":"txn-id"}');

COMMIT;
```

## Hold Flow (Card Authorization)

```sql
-- Hold qo'yish
BEGIN;
SELECT * FROM accounts WHERE id = $1 FOR UPDATE;
UPDATE accounts SET
    available_balance = available_balance - 50000,
    hold_amount = hold_amount + 50000,
    version = version + 1
WHERE id = $1 AND version = $2;
-- Event: HoldPlacedEvent
COMMIT;

-- Hold capture (partial: 30000 of 50000)
BEGIN;
UPDATE accounts SET
    balance_minor = balance_minor - 30000,
    hold_amount = hold_amount - 50000,
    available_balance = available_balance + 20000, -- qoldiq qaytarildi
    version = version + 1
WHERE id = $1;
COMMIT;
```

## Event Sourcing Flow

```
WRITE (Command):
  1. Load: snapshot(v100) + events(101-105) → Account state
  2. Apply: Account.Debit() → new event(v106)
  3. Save: INSERT INTO event_store (version=106)
  4. Snapshot: if version % 100 == 0 → INSERT INTO snapshots

READ (Query):
  Event Store ──publish──→ Projection Handler ──update──→ Read Model Table
  │                                                        │
  └── Source of truth                                      └── Denormalized, fast
```

## Partitioning Flow

```
INSERT INTO ledger_entries (..., created_at='2026-03-15')
  │
  └── PostgreSQL auto-routes → ledger_entries_2026_03 (March partition)

-- pg_cron har oy 25-da:
SELECT create_next_month_partitions();
  → ledger_entries_2026_04
  → audit_log_2026_04
  → event_store_2026_04
  → transfers_2026_04
```

## Backup Flow

```
Continuous:
  PostgreSQL WAL ──stream──→ Replica (real-time)
                ──archive──→ /backups/wal/ (PITR)

Daily (2:00 AM):
  pg_basebackup ──→ /backups/daily/2026-03-30.tar.gz

Recovery:
  RPO < 1 daqiqa (WAL stream)
  RTO < 15 daqiqa (restore + replay WAL)
```

## Reconciliation Flow (Daily 3:00 AM)

```sql
-- pg_cron job:
SELECT run_daily_reconciliation();

-- 1. SUM(debit) = SUM(credit)?
-- 2. account.balance = SUM(ledger_entries)?
-- 3. account.balance = replay(event_store)?
-- Mismatch → INSERT INTO audit_log (action='RECONCILIATION_MISMATCH')
-- Mismatch → notification → admin alert
```

## pg_cron Jobs

```
Every 5 min:   Clean expired OTP/challenges
Every 1 hour:  Clean expired sessions
Every 1 day:   Clean expired idempotency keys
Every 1 day:   Run reconciliation (3:00 AM)
Every 1 month: Create next month partitions (25th)
```

## Query Performance

```sql
-- pg_stat_statements: top slow queries
-- EXPLAIN (ANALYZE, BUFFERS) for every new query
-- Partial indexes: WHERE status = 'ACTIVE'
-- Covering indexes: INCLUDE (balance, version)
-- Connection: PgBouncer transaction mode
```
