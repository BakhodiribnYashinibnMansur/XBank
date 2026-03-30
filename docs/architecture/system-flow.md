# XBank — Umumiy Ishlash Flow

## Tizim Arxitekturasi

```
                    ┌──────────────┐
                    │   Browser    │
                    │  (Test UI)   │
                    └──────┬───────┘
                           │ HTTPS
                           ▼
                    ┌──────────────┐
                    │  GoFiber v2  │
                    │  (Port 8080) │
                    │  14 Middleware│
                    └──────┬───────┘
                           │
              ┌────────────┼────────────┐
              │            │            │
              ▼            ▼            ▼
        ┌──────────┐ ┌──────────┐ ┌──────────┐
        │ Identity │ │ Account  │ │ Transfer │ ...
        │ Context  │ │ Context  │ │ Context  │
        └────┬─────┘ └────┬─────┘ └────┬─────┘
             │            │            │
             └────────────┼────────────┘
                          │
              ┌───────────┼───────────┐
              │           │           │
              ▼           ▼           ▼
        ┌──────────┐ ┌──────────┐ ┌──────────┐
        │PostgreSQL│ │  Redis   │ │ EventBus │
        │(via PgB) │ │          │ │          │
        └──────────┘ └──────────┘ └──────────┘
              │           │           │
              ▼           │           ▼
        ┌──────────┐     │     ┌──────────┐
        │ Replica  │     │     │  Redis   │
        │ (Read)   │     │     │ Streams  │
        └──────────┘     │     │ (Queue)  │
                         │     └──────────┘
                         ▼
                   ┌──────────┐
                   │Prometheus│──→ Grafana
                   └──────────┘
```

## Foydalanuvchi Ro'yxatdan O'tish Flow

```
User                    Frontend               Backend                 PostgreSQL          Redis
 │                         │                      │                       │                  │
 ├── Register form ──────→ │                      │                       │                  │
 │                         ├── POST /auth/register│                       │                  │
 │                         │                      ├── Validate input      │                  │
 │                         │                      ├── bcrypt(password)    │                  │
 │                         │                      ├── INSERT INTO users ─→│                  │
 │                         │                      ├── Event: UserRegistered                  │
 │                         │                      ├── Generate JWT + refresh                 │
 │                         │                      ├── INSERT INTO sessions│                  │
 │                         │                      ├── SET session ───────────────────────────→│
 │                         │ ←── {access, refresh} │                       │                  │
 │ ←── Dashboard ─────── │                      │                       │                  │
```

## Login + 2FA Flow

```
User                    Frontend               Backend                 Redis
 │                         │                      │                       │
 ├── Email + Password ───→ │                      │                       │
 │                         ├── POST /auth/login ─→│                       │
 │                         │                      ├── Verify password     │
 │                         │                      ├── 2FA enabled?        │
 │                         │ ←── needs_2fa: true   │                       │
 ├── TOTP Code ──────────→ │                      │                       │
 │                         ├── POST /auth/2fa ───→│                       │
 │                         │                      ├── Verify TOTP code    │
 │                         │                      ├── Device fingerprint  │
 │                         │                      ├── Create session ────────────────────────→│
 │                         │ ←── {access, refresh} │                       │
 │ ←── Dashboard ─────── │                      │                       │
```

## Transfer To'liq Flow

```
User                Frontend            Backend                    PostgreSQL         Redis
 │                     │                   │                          │                 │
 ├── Transfer form ──→ │                   │                          │                 │
 │   (beneficiary,     │                   │                          │                 │
 │    amount,          │                   │                          │                 │
 │    currency)        │                   │                          │                 │
 │                     ├── Confirmation ──→ User confirms             │                 │
 │ ←── TOTP? ───────── │                   │                          │                 │
 ├── TOTP code ──────→ │                   │                          │                 │
 │                     ├── POST /transfers │                          │                 │
 │                     │   + Idempotency   │                          │                 │
 │                     │   + X-Signature   │                          │                 │
 │                     │                   ├── 1. Idempotency check ────────────────────→│
 │                     │                   │       (Redis fast check)  │                 │
 │                     │                   ├── 2. BEGIN SERIALIZABLE ─→│                 │
 │                     │                   ├── 3. Fraud Check          │                 │
 │                     │                   │       velocity + device ──────────────────→│
 │                     │                   ├── 4. AML Screening       │                 │
 │                     │                   ├── 5. Lock accounts ─────→│ (FOR UPDATE)    │
 │                     │                   ├── 6. Check balance ──────→│                 │
 │                     │                   ├── 7. Debit source ──────→│                 │
 │                     │                   ├── 8. Credit target ─────→│                 │
 │                     │                   ├── 9. Ledger entries ────→│                 │
 │                     │                   ├── 10. Save transfer ────→│                 │
 │                     │                   ├── 11. COMMIT ───────────→│                 │
 │                     │                   ├── 12. Publish events ────────────────────→│
 │                     │                   │       (Redis Pub/Sub)     │                 │
 │                     │ ←── 201 Created    │                          │                 │
 │                     │                   │                          │                 │
 │ ←── SSE: "Transfer  │ ←── SSE event ─── │ ←── Redis Pub/Sub ─────────────────────── │
 │     muvaffaqiyatli!" │                   │                          │                 │
```

## Monitoring Flow

```
Request ──→ Fiber ──→ Prometheus Middleware ──→ /metrics endpoint
                                                     │
                                               Prometheus scrape (15s)
                                                     │
                                               Grafana Dashboard
                                                     │
                                    ┌────────────────┼────────────────┐
                                    │                │                │
                              Request Rate     Error Rate      Latency P99
                                    │                │                │
                              Alert Rules ──→ Notification (threshold oshsa)
```

## Nightly Jobs Flow

```
03:00 AM (pg_cron):
  ├── Reconciliation
  │     ├── SUM(debit) == SUM(credit)? ✓/✗
  │     ├── account.balance == ledger sum? ✓/✗
  │     ├── account.balance == event replay? ✓/✗
  │     └── Mismatch → ALERT
  │
  ├── Clean expired sessions
  ├── Clean expired idempotency keys
  └── Run scheduled transfers (DAILY, WEEKLY, MONTHLY)

25th of month (pg_cron):
  └── Create next month partitions
```
