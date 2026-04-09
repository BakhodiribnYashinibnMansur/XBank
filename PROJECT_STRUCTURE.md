# XBank — Project Structure

## Architecture

DDD Modular Monolith with Event Sourcing, CQRS, and Saga pattern.

```
┌─────────────────────────────────────────────────┐
│                   HTTP Layer                     │
│         Fiber handlers + 14 middleware           │
├─────────────────────────────────────────────────┤
│              Application Layer                   │
│        CQRS Command/Query handlers               │
├─────────────────────────────────────────────────┤
│               Domain Layer                       │
│    Aggregates, Events, Specifications            │
│         (ZERO external imports)                  │
├─────────────────────────────────────────────────┤
│           Infrastructure Layer                   │
│    PostgreSQL, Redis, Kafka, Vault               │
└─────────────────────────────────────────────────┘
```

## Directory Layout

```
cmd/
  app/                   → Application entry point

internal/
  app/                   → App bootstrap, route registration
  kernel/                → Shared kernel (cross-cutting concerns)
    domain/              → Shared domain types (Money, Currency, errors)
    application/         → Event bus, shared interfaces
    infrastructure/
      config/            → Viper-based configuration
      db/
        postgres/        → pgxpool, TxManager, DBTX pattern
        redis/           → Session cache, login limiter, challenge cache
        mongodb/         → Audit log persistence
      kafka/             → Producer, consumer, DLQ, schema registry
      middleware/        → 14-middleware stack
      security/
        jwt/             → ES256 JWT service + TOTP
        crypto/          → AES encryption, HMAC signing, tokenization
        vault/           → HashiCorp Vault integration
      httpclient/        → Instrumented HTTP client + SSRF guard
      metrics/           → Prometheus metrics
      tracing/           → OpenTelemetry → Jaeger
      sse/               → Server-Sent Events hub
    outbox/              → Transactional outbox pattern

  context/               → Bounded Contexts (DDD)
    banking/
      core/
        account/         → Event-sourced bank account (CQRS)
        card/            → Card issuance, PIN, tokenization, 3DS
        transfer/        → Event-sourced money transfers
        ledger/          → Double-entry bookkeeping
      supporting/
        beneficiary/     → Trusted transfer recipients
        exchange/        → Currency exchange rates
        fraud/           → AML/fraud risk scoring
        kyc/             → KYC document verification
        reconciliation/  → Balance reconciliation engine

    iam/
      core/
        challenge/       → Step-up authentication (MFA)
      generic/
        user/            → User registration, profile
        session/         → JWT sessions, TOTP, refresh tokens
        authz/           → RBAC (roles, permissions, policies)
        usersetting/     → Per-user key-value settings
      supporting/
        audit/           → Audit logs + endpoint history
        contact/         → User-to-user contacts
        device/          → Device fingerprinting

    admin/
      generic/
        featureflag/     → Feature flags with rollout %
      supporting/
        sitesetting/     → Global site settings
        statistics/      → Daily snapshot statistics
        dataexport/      → GDPR data export
        integration/     → Third-party API registry

    content/
      generic/
        notification/    → User notifications
        translation/     → i18n translations (uz/ru/en)
        file/            → File metadata (MinIO)
      supporting/
        announcement/    → Multi-language announcements

    ops/
      generic/
        systemerror/     → System error tracking
        ratelimit/       → Rate limit rules CRUD
        metric/          → Application metrics
      supporting/
        errorcode/       → Error code registry
        iprule/          → IP-based access rules

migrations/              → Goose SQL migrations (001-035)
proto/                   → Protobuf definitions (Kafka messages)
test/
  arch/                  → Architecture fitness tests
  integration/           → Testcontainers integration tests
  e2e/                   → End-to-end API flow tests
  performance/           → Stress, spike, endurance, breakpoint tests
tests/
  load/                  → HTTP load tests
docs/
  swagger/               → Swagger/OpenAPI docs
scripts/
  protogen.sh            → Protobuf code generation
```

## Bounded Context Structure

Each BC follows this structure:

```
bc-name/
  domain/
    aggregate.go         → Entity, factory, business rules
    events.go            → Domain events
  application/
    command/
      service.go         → Command handlers (write side)
  infrastructure/
    postgres/
      write_repo.go      → PostgreSQL repository
      read_repo.go       → Read-optimized queries (optional)
  interfaces/
    http/
      handler.go         → Fiber HTTP handlers
      request.go         → Request DTOs
      response.go        → Response DTOs
  bc.go                  → Wiring: creates service + handler from pool
```

## BC Tier Classification

| Tier | Description | Examples |
|------|-------------|---------|
| **Core** | Primary business value, complex domain logic | Account, Transfer, Card, Challenge |
| **Generic** | Reusable across domains, standard patterns | User, Session, FeatureFlag, Notification |
| **Supporting** | Assists core BCs, simpler logic | KYC, Fraud, Audit, ErrorCode |

## Key Patterns

- **Event Sourcing**: Account + Transfer aggregates (EAV event store)
- **CQRS**: Separate write (events) and read (projections) models
- **Transactional Outbox**: Events published within DB transaction
- **Saga**: Cross-BC operations via Kafka events
- **Shared Kernel**: `internal/kernel/domain/` for Money, Currency, TxManager
- **DBTX Pattern**: `ExtractDBTX(ctx, pool)` — transparent transaction propagation
