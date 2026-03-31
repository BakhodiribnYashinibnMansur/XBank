# Backend Operation Flow

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
│ 12. 2FA (if required)                        │
│ 13. KYC Required (if required)               │
│ 14. Device Fingerprint                       │
└─────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────┐
│              HTTP HANDLER                    │
│  - Request DTO parse + validate              │
│  - Create Command/Query                      │
│  - Call application service                  │
│  - Response DTO format                       │
└─────────────────────────────────────────────┘
    │
    ▼
┌─────────────────────────────────────────────┐
│         APPLICATION LAYER (Use Case)         │
│  Command Handler:                            │
│  1. Idempotency check                        │
│  2. UnitOfWork.Begin(SERIALIZABLE)           │
│  3. Repository.Load (load aggregate)         │
│  4. Domain logic (aggregate methods)         │
│  5. Specification check (balance, limit)     │
│  6. Repository.Save (event store)            │
│  7. UnitOfWork.Commit                        │
│  8. EventBus.Publish (after commit)          │
│                                              │
│  Query Handler:                              │
│  1. Query from Read Model (read replica)     │
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
│  - Redis (session, cache, pub/sub)            │
│  - Kafka (async message queue, Protobuf)     │
│  - Crypto (AES, HMAC, bcrypt)                │
│  - JWT (ES256)                               │
│  - Event Bus (sync + async)                  │
└─────────────────────────────────────────────┘

## Transfer Flow (detailed)

```
POST /api/v1/transfers
    │
    ▼
Middleware Stack (14 total) → Handler
    │
    ▼
InitiateTransferHandler:
    │
    ├── 1. Idempotency check (DB lookup)
    │       ├── Key exists → return cached response
    │       └── Key missing → continue
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
            └── → Kafka (async, Protobuf)
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
    └── Async: Kafka Producer (Protobuf serialize)
            │
            ├── xbank.transfers.created (key=account_id)
            │       ├── Consumer Group: fraud-group → FraudAnalysis
            │       └── Consumer Group: notification-group → StatementGenerator
            │
            └── xbank.transfers.failed (key=transfer_id)
                    └── Consumer Group: alert-group → AlertService
                            │
                            └── Retry 3x → fail → xbank.dlq topic
                                                    │
                                                    └── Admin manual review
```

## Request/Response Logging (to DB)

<!-- Middleware #9 (Audit Logger) — saves every HTTP request and response
     to PostgreSQL. Required by banking regulators.
     Sensitive data is MASKED. -->

### Request Log Table

```sql
CREATE TABLE request_logs (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Request
    method           VARCHAR(10) NOT NULL,         -- GET, POST, PUT, DELETE
    path             VARCHAR(500) NOT NULL,         -- /api/v1/transfers
    query_params     TEXT,                          -- ?page=1&limit=20
    request_headers  JSONB,                         -- masked headers
    request_body     JSONB,                         -- masked body

    -- Response
    status_code      SMALLINT NOT NULL,             -- 200, 400, 401, 500
    response_body    JSONB,                         -- masked body
    response_headers JSONB,

    -- Metadata
    request_id       VARCHAR(64) NOT NULL,          -- X-Request-ID
    correlation_id   VARCHAR(64),                   -- X-Correlation-ID
    user_id          UUID,                          -- from JWT (if authenticated)
    session_id       UUID,                          -- Session ID
    ip_address       INET NOT NULL,                 -- Client IP
    user_agent       TEXT,                          -- User-Agent header
    device_id        VARCHAR(255),                  -- X-Device-Fingerprint

    -- Performance
    duration_ms      INTEGER NOT NULL,              -- Response time (ms)
    request_size     INTEGER,                       -- Request body size (bytes)
    response_size    INTEGER,                       -- Response body size (bytes)

    -- Timestamps
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
) PARTITION BY RANGE (created_at);

-- Monthly partitions
CREATE TABLE request_logs_2026_03 PARTITION OF request_logs
    FOR VALUES FROM ('2026-03-01') TO ('2026-04-01');
CREATE TABLE request_logs_2026_04 PARTITION OF request_logs
    FOR VALUES FROM ('2026-04-01') TO ('2026-05-01');

-- Indexes
CREATE INDEX idx_req_logs_user      ON request_logs (user_id, created_at DESC);
CREATE INDEX idx_req_logs_path      ON request_logs (path, created_at DESC);
CREATE INDEX idx_req_logs_status    ON request_logs (status_code, created_at DESC);
CREATE INDEX idx_req_logs_request   ON request_logs (request_id);
CREATE INDEX idx_req_logs_corr      ON request_logs (correlation_id);
CREATE INDEX idx_req_logs_created   ON request_logs (created_at DESC);
```

### Middleware Flow

```
Request arrives
    │
    ▼
┌─ Audit Logger Middleware (#9) ──────────────────────────────────────┐
│                                                                      │
│  1. When REQUEST starts:                                             │
│     start_time = time.Now()                                         │
│     request_id = c.Locals("requestID")                              │
│     request_body = c.Body()          ← clone it (stream can only   │
│                                         be read once)               │
│                                                                      │
│  2. c.Next() → handler executes → response ready                    │
│                                                                      │
│  3. When RESPONSE is ready:                                          │
│     duration = time.Since(start_time)                               │
│     status_code = c.Response().StatusCode()                         │
│     response_body = c.Response().Body()                             │
│                                                                      │
│  4. MASK sensitive data (AFTER response is sent)                    │
│                                                                      │
│  5. Async DB write (we DON'T WAIT for the response)                │
│     go func() {                                                      │
│         db.InsertRequestLog(logEntry)                               │
│     }()                                                              │
│                                                                      │
│  ⚠️ DB write is ASYNC — does not affect response speed               │
└──────────────────────────────────────────────────────────────────────┘
```

### Sensitive Data Masking

```
Fields that are NEVER logged (completely removed):
  ❌ password           → "[REDACTED]"
  ❌ encrypted_pan      → "[REDACTED]"
  ❌ encrypted_pin      → "[REDACTED]"
  ❌ encrypted_cvv      → "[REDACTED]"
  ❌ totp_secret        → "[REDACTED]"
  ❌ refresh_token      → "[REDACTED]"
  ❌ Authorization header → "Bearer [REDACTED]"

Masked fields (partially shown):
  ⚠️ email             → "b***@example.com"
  ⚠️ phone             → "+998***1234"
  ⚠️ card_number       → "****7890"
  ⚠️ ip_address        → stored in full (needed for security)

Go pseudocode:
  func maskRequestBody(body []byte) []byte {
      var data map[string]interface{}
      json.Unmarshal(body, &data)

      redactFields := []string{
          "password", "encrypted_pan", "encrypted_pin",
          "encrypted_cvv", "totp_secret", "refresh_token",
      }
      for _, field := range redactFields {
          if _, exists := data[field]; exists {
              data[field] = "[REDACTED]"
          }
      }

      // Nested encrypted objects
      for key, val := range data {
          if m, ok := val.(map[string]interface{}); ok {
              if _, hasCipher := m["ciphertext"]; hasCipher {
                  data[key] = "[E2EE_REDACTED]"
              }
          }
      }

      masked, _ := json.Marshal(data)
      return masked
  }
```

### What Gets Logged (Examples)

```
✅ Login (successful):
  method: POST, path: /api/v1/auth/login, status: 200
  request_body:  { "email": "b***@example.com", "password": "[REDACTED]" }
  response_body: { "status": "success", "data": { "access_token": "[REDACTED]" } }
  user_id: null → user-uuid (after login)
  duration_ms: 145

✅ Transfer:
  method: POST, path: /api/v1/transfers, status: 201
  request_body:  { "from_account": "acc-1", "to_account": "acc-2", "amount": 100000 }
  response_body: { "status": "success", "data": { "transfer_id": "txn-uuid" } }
  user_id: user-uuid
  duration_ms: 320

✅ Card creation (E2EE):
  method: POST, path: /api/v1/cards, status: 201
  request_body:  {
    "account_id": "acc-uuid",
    "cardholder_name": "BAKHODIR YASHINI",
    "encrypted_pan": "[E2EE_REDACTED]",
    "encrypted_pin": "[E2EE_REDACTED]",
    "encrypted_cvv": "[E2EE_REDACTED]"
  }
  response_body: { "status": "success", "data": { "card_id": "card-uuid", "last_four": "7890" } }
  duration_ms: 520

❌ Unauthorized:
  method: GET, path: /api/v1/accounts, status: 401
  request_body:  null
  response_body: { "status": "error", "error": { "code": "UNAUTHORIZED" } }
  user_id: null
  duration_ms: 12

❌ Rate limited:
  method: POST, path: /api/v1/auth/login, status: 429
  request_body:  { "email": "b***@example.com", "password": "[REDACTED]" }
  response_body: { "status": "error", "error": { "code": "RATE_LIMITED" } }
  duration_ms: 3
```

### Log Retention Period

```
Rules:
  - Request logs:   90 days (active), then archived
  - Archive:        2 years (cold storage / compressed)
  - Audit log:      7 years (regulatory requirement)

pg_cron (weekly cleanup):
  SELECT cron.schedule('drop-old-request-logs', '0 3 * * 0',
    $$DROP TABLE IF EXISTS request_logs_old_partition$$
  );

Partition size (approximate):
  - 1000 req/hour × 24 × 30 = ~720,000 rows/month
  - Average row size: ~2 KB
  - Monthly partition: ~1.4 GB
  - 90 days: ~4.2 GB
```

### Admin API (Log search)

```
GET /api/v1/admin/request-logs?user_id=xxx&path=/api/v1/transfers&status=500
Authorization: Bearer <admin-JWT>

Filters:
  - user_id         — specific user
  - path            — specific endpoint
  - method          — GET, POST, PUT, DELETE
  - status_code     — 200, 400, 500 etc.
  - ip_address      — specific IP
  - correlation_id  — single request chain
  - date_from/to    — time range
  - min_duration_ms — slow requests (> 1000ms)

Response:
{
  "status": "success",
  "data": [...],
  "meta": {
    "total": 1523,
    "page": 1,
    "limit": 50,
    "cursor": "next-cursor..."
  }
}
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
