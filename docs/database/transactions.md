# Transaction Layer

## ACID
```
Atomicity    — Transfer: debit A + credit B = one atomic operation
Consistency  — CHECK(balance >= 0), FK, domain invariants
Isolation    — SERIALIZABLE for critical transfers
Durability   — WAL + fsync + replication
```

## Isolation Levels
```sql
READ UNCOMMITTED   — NEVER (dirty reads!)
READ COMMITTED     — Default, for most operations
REPEATABLE READ    — Statement, reconciliation
SERIALIZABLE       — Transfer (safest)
```

### Where to use which:
| Operation | Isolation | Reason |
|---|---|---|
| Transfer | SERIALIZABLE | Write skew, phantom read protection |
| Balance check + Debit | SERIALIZABLE | Concurrent debit protection |
| Statement | REPEATABLE READ | Consistent snapshot |
| Simple CRUD | READ COMMITTED | Default |

## 6 Concurrency Problems

| Problem | Description | Protection |
|---|---|---|
| Dirty Read | Reading uncommitted data | READ COMMITTED |
| Dirty Write | Uncommitted overwrite | PG all levels |
| Non-repeatable Read | Different result on re-read | REPEATABLE READ |
| Phantom Read | New rows appear | SERIALIZABLE |
| Lost Update | Concurrent update lost | Optimistic/Pessimistic lock |
| Write Skew | Concurrent read → invalid write | SERIALIZABLE |

## Locking Strategies

### Optimistic Locking (version)
```sql
UPDATE accounts SET balance_minor=$1, version=version+1
WHERE id=$2 AND version=$3;
-- 0 rows → retry
```

### Pessimistic Lock (FOR UPDATE)
```sql
SELECT * FROM accounts WHERE id=$1 FOR UPDATE;
-- other transactions wait
```

### FOR UPDATE SKIP LOCKED (parallel workers)
```sql
SELECT * FROM transfers WHERE status='PENDING'
ORDER BY created_at LIMIT 10 FOR UPDATE SKIP LOCKED;
```

### Advisory Locks
```sql
SELECT pg_advisory_xact_lock(hashtext('account:' || $1::text));
-- Automatically unlocked when transaction ends
```

### Lock Timeout Configuration
<!-- Limit lock wait time — to prevent deadlocks and improve performance -->
```sql
-- Session level (per connection)
SET lock_timeout = '5s';           -- waits 5 seconds, then raises error
SET statement_timeout = '30s';     -- cancels query if it exceeds 30 seconds

-- Transaction level (per single transaction)
SET LOCAL lock_timeout = '3s';
SET LOCAL statement_timeout = '10s';

-- In Go:
-- Default values are set via pgx config
-- Separate timeouts can be configured for each use case
```

## Retry Strategy
```go
// Retryable: 40001 (serialization_failure), 40P01 (deadlock)
// Not retryable: 23514 (check_violation), 23505 (unique_violation)
// Backoff: 100ms → 200ms → 400ms (max 3)
```

## Error Classification
| PG Code | Error | Retry? |
|---|---|---|
| 40001 | serialization_failure | Yes |
| 40P01 | deadlock_detected | Yes |
| 08006 | connection_failure | Yes |
| 23514 | check_violation | No (business error) |
| 23505 | unique_violation | No (idempotency) |
| 23503 | foreign_key_violation | No (not found) |
