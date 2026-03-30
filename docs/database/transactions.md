# Transaction Layer

## ACID
```
Atomicity    — Transfer: debit A + credit B = bitta atomic operation
Consistency  — CHECK(balance >= 0), FK, domain invariants
Isolation    — SERIALIZABLE muhim transfers uchun
Durability   — WAL + fsync + replication
```

## Isolation Levels
```sql
READ UNCOMMITTED   — HECH QACHON (dirty reads!)
READ COMMITTED     — Default, ko'p operatsiyalar uchun
REPEATABLE READ    — Statement, reconciliation
SERIALIZABLE       — Transfer (eng xavfsiz)
```

### Qayerda qaysi:
| Operatsiya | Isolation | Sabab |
|---|---|---|
| Transfer | SERIALIZABLE | Write skew, phantom read himoya |
| Balance check + Debit | SERIALIZABLE | Concurrent debit himoya |
| Statement | REPEATABLE READ | Consistent snapshot |
| Oddiy CRUD | READ COMMITTED | Default |

## 6 ta Concurrency Muammo

| Muammo | Tavsif | Himoya |
|---|---|---|
| Dirty Read | Uncommitted o'qish | READ COMMITTED |
| Dirty Write | Uncommitted overwrite | PG barcha darajalar |
| Non-repeatable Read | Qayta o'qishda farq | REPEATABLE READ |
| Phantom Read | Yangi qatorlar | SERIALIZABLE |
| Lost Update | Concurrent update yo'qolish | Optimistic/Pessimistic lock |
| Write Skew | Concurrent read → invalid write | SERIALIZABLE |

## Locking Strategiyalari

### Optimistic Locking (version)
```sql
UPDATE accounts SET balance_minor=$1, version=version+1
WHERE id=$2 AND version=$3;
-- 0 rows → retry
```

### Pessimistic Lock (FOR UPDATE)
```sql
SELECT * FROM accounts WHERE id=$1 FOR UPDATE;
-- boshqa tranzaksiyalar kutadi
```

### FOR UPDATE SKIP LOCKED (parallel workers)
```sql
SELECT * FROM transfers WHERE status='PENDING'
ORDER BY created_at LIMIT 10 FOR UPDATE SKIP LOCKED;
```

### Advisory Locks
```sql
SELECT pg_advisory_xact_lock(hashtext('account:' || $1::text));
-- Tranzaksiya tugagach avtomatik unlock
```

### Lock Timeout Konfiguratsiyasi
<!-- Lock kutish vaqtini cheklash — deadlock oldini olish va performance uchun -->
```sql
-- Session level (connection uchun)
SET lock_timeout = '5s';           -- 5 sekund kutib, keyin xato beradi
SET statement_timeout = '30s';     -- query 30 sekund dan oshsa bekor qilish

-- Transaction level (bitta tranzaksiya uchun)
SET LOCAL lock_timeout = '3s';
SET LOCAL statement_timeout = '10s';

-- Go da:
-- pgx config orqali default qiymatlar o'rnatiladi
-- Har bir use case uchun alohida timeout belgilash mumkin
```

## Retry Strategy
```go
// Retryable: 40001 (serialization_failure), 40P01 (deadlock)
// Not retryable: 23514 (check_violation), 23505 (unique_violation)
// Backoff: 100ms → 200ms → 400ms (max 3)
```

## Error Classification
| PG Code | Xato | Retry? |
|---|---|---|
| 40001 | serialization_failure | Yes |
| 40P01 | deadlock_detected | Yes |
| 08006 | connection_failure | Yes |
| 23514 | check_violation | No (business error) |
| 23505 | unique_violation | No (idempotency) |
| 23503 | foreign_key_violation | No (not found) |
