# Backend Ishlash Flow

## Request Lifecycle

```
Client Request
    │
    ▼
┌─────────────────────────────────────────────┐
│              FIBER MIDDLEWARE STACK           │
│                                              │
│  1. Recovery (panic catch)                   │
│  2. RequestID (unique ID)                    │
│  3. Correlation-ID (X-Correlation-ID)        │
│  4. Helmet (security headers)                │
│  5. CORS                                     │
│  6. CSRF token check                         │
│  7. Rate Limiter (Redis sliding window)      │
│  8. Prometheus Metrics                       │
│  9. Audit Logger (request log)               │
│ 10. Session (JWT validate → Redis check)     │
│ 11. RBAC/ABAC (role + attribute check)       │
│ 12. 2FA (agar kerak)                         │
│ 13. KYC Required (agar kerak)                │
│ 14. Device Fingerprint                       │
└─────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────┐
│              HTTP HANDLER                    │
│  - Request DTO parse + validate              │
│  - Command/Query yaratish                    │
│  - Application service chaqirish             │
│  - Response DTO format                       │
└─────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────┐
│         APPLICATION LAYER (Use Case)         │
│  Command Handler:                            │
│  1. Idempotency check                        │
│  2. UnitOfWork.Begin(SERIALIZABLE)           │
│  3. Repository.Load (aggregate yuklash)      │
│  4. Domain logic (aggregate methods)         │
│  5. Specification check (balance, limit)     │
│  6. Repository.Save (event store)            │
│  7. UnitOfWork.Commit                        │
│  8. EventBus.Publish (after commit)          │
│                                              │
│  Query Handler:                              │
│  1. Read Model'dan query (read replica)      │
│  2. Cursor-based pagination                  │
└─────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────┐
│            DOMAIN LAYER                      │
│  - Aggregate Root (business logic)           │
│  - Value Objects (Money, Currency)           │
│  - Domain Events (state changes)             │
│  - Specifications (business rules)           │
│  - Domain Services (cross-aggregate logic)   │
│  - ZERO external dependencies                │
└─────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────┐
│         INFRASTRUCTURE LAYER                 │
│  - PostgreSQL (pgx, event store, snapshots)  │
│  - Redis (session, cache, queue, pub/sub)    │
│  - Crypto (AES, HMAC, bcrypt)                │
│  - JWT (ES256)                               │
│  - Event Bus (sync + async)                  │
└─────────────────────────────────────────────┘

## Transfer Flow (batafsil)

```
POST /api/v1/transfers
    │
    ▼
Middleware Stack (14 ta) → Handler
    │
    ▼
InitiateTransferHandler:
    │
    ├── 1. Idempotency check (DB lookup)
    │       ├── Key bor → return cached response
    │       └── Key yo'q → davom et
    │
    ├── 2. BEGIN SERIALIZABLE transaction
    │
    ├── 3. Fraud Check
    │       ├── Velocity (5+/hour?)
    │       ├── Device match?
    │       ├── Amount pattern?
    │       └── Risk score → LOW/MEDIUM/HIGH
    │           ├── LOW → continue
    │           ├── MEDIUM → require 2FA
    │           └── HIGH → BLOCK + alert
    │
    ├── 4. AML Screening
    │       ├── Amount > $10K? → FLAG
    │       ├── New account + large amount? → FLAG
    │       └── Risk > 70 → BLOCK
    │
    ├── 5. Lock accounts (UUID order → deadlock prevention)
    │       SELECT ... FOR UPDATE
    │
    ├── 6. Validate balance
    │       ├── SufficientBalanceSpec
    │       ├── DailyLimitSpec
    │       └── MonthlyLimitSpec
    │
    ├── 7. Execute double-entry (TransferDomainService)
    │       ├── sourceAccount.Debit(amount)    → AccountDebitedEvent
    │       ├── targetAccount.Credit(amount)   → AccountCreditedEvent
    │       └── transfer.MarkCompleted()       → TransferCompletedEvent
    │
    ├── 8. HMAC sign transaction
    │
    ├── 9. Save (event store + ledger + transfer)
    │
    ├── 10. COMMIT
    │
    └── 11. Publish events (after commit)
            ├── → Notification service (SSE push)
            ├── → Audit log
            └── → Redis Streams (async)
```

## Account Load Flow (Event Sourcing)

```
AccountRepository.FindByID(id):
    │
    ├── 1. SnapshotStore.LoadLatest(id)
    │       → snapshot (version=100, state JSON)
    │
    ├── 2. EventStore.LoadAfter(id, version=100)
    │       → events [101, 102, 103, ...]
    │
    ├── 3. Account.RebuildFromSnapshot(snapshot)
    │
    ├── 4. Account.Replay(events)
    │       ├── event 101: AccountCreditedEvent → balance += 500
    │       ├── event 102: AccountDebitedEvent → balance -= 200
    │       └── event 103: HoldPlacedEvent → available -= 100
    │
    └── 5. Return Account (current state)
```

## Async Processing Flow

```
Domain Event (after commit)
    │
    ├── Sync: EventBus.Publish()
    │       └── NotificationHandler → SSE push
    │
    └── Async: Redis Streams
            │
            ├── xbank:transfers:created
            │       ├── Consumer: FraudAnalysis (background deep scan)
            │       └── Consumer: StatementGenerator
            │
            └── xbank:transfers:failed
                    └── Consumer: AlertService
                            │
                            └── Retry 3x → fail → DLQ
                                                    │
                                                    └── Admin manual review
```

## Error Handling Flow

```
Error occurs
    │
    ├── Retryable? (40001, 40P01, 08006)
    │       ├── Yes → Retry (100ms → 200ms → 400ms, max 3)
    │       │       └── Still fail → DLQ
    │       └── No → Business error
    │               ├── 23514 → ErrInsufficientFunds → 400
    │               ├── 23505 → Idempotency duplicate → return cached
    │               └── Domain error → mapped HTTP status
    │
    └── Response:
        {
          "status": "error",
          "error": {"code": "...", "message": "..."},
          "meta": {"request_id": "...", "correlation_id": "..."}
        }
```
