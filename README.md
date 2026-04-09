# XBank

A banking application built with Go, following DDD Modular Monolith architecture with Event Sourcing, CQRS, and Saga patterns.

## Tech Stack

| Component | Technology |
|-----------|-----------|
| **Language** | Go 1.25+ |
| **HTTP** | GoFiber v2 |
| **Database** | PostgreSQL 17 (primary + read replica) |
| **Cache** | Redis 7 (sessions, rate limiting, pub/sub) |
| **Queue** | Apache Kafka (Protobuf serialization) |
| **Secrets** | HashiCorp Vault |
| **Observability** | Prometheus + Jaeger (OpenTelemetry) |
| **Containers** | Docker + docker-compose |

## Quick Start

```bash
# 1. Clone and setup
git clone https://github.com/BakhodiribnYashinibnMansur/XBank.git
cd XBank

# 2. Generate JWT keys
make jwt-keys

# 3. Start with Docker
make docker-up

# 4. Or run locally (requires PostgreSQL, Redis)
cp config.yml config.local.yml  # edit as needed
make run
```

## Development

```bash
# Hot reload
make dev

# Run all tests
make test

# Architecture tests
make test-arch

# Integration tests (Docker required)
make test-integration

# E2E tests (Docker required)
make test-e2e

# Performance tests (server must be running)
make test-stress
make test-spike
make test-endurance
make test-breakpoint

# Linting
make lint

# Benchmarks
make bench
```

## Migrations

```bash
make migrate-up                     # Apply all
make migrate-down                   # Rollback last
make migrate-status                 # View status
make migrate-create name=add_cards  # Create new
```

## Architecture

31 Bounded Contexts organized by domain and tier:

| Domain | Core | Generic | Supporting |
|--------|------|---------|------------|
| **Banking** | Account, Card, Transfer, Ledger | — | Beneficiary, Exchange, Fraud, KYC, Reconciliation |
| **IAM** | Challenge | User, Session, Authz, UserSetting | Audit, Contact, Device |
| **Admin** | — | FeatureFlag | SiteSetting, Statistics, DataExport, Integration |
| **Content** | — | Notification, Translation, File | Announcement |
| **Ops** | — | SystemError, RateLimit, Metric | ErrorCode, IPRule |

See [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md) for detailed directory layout.

## API Endpoints

| Group | Prefix | Auth | Description |
|-------|--------|------|-------------|
| **Auth** | `/api/v1/auth/` | Public | Register, Login, Refresh, Logout, TOTP |
| **Accounts** | `/api/v1/accounts/` | JWT | Create, Deposit, Withdraw, Close |
| **Transfers** | `/api/v1/transfers/` | JWT + HMAC | Send, Schedule, History |
| **Cards** | `/api/v1/cards/` | JWT + HMAC | Issue, Activate, Tokenize, 3DS |
| **Admin** | `/api/v1/admin/` | JWT + Admin | KYC review, Fraud, RBAC, Settings |

Full API docs available at `/swagger/` and `/docs` when server is running.

## Configuration

Configuration uses [Viper](https://github.com/spf13/viper) with layered sources:

1. **Defaults** — sensible development defaults
2. **YAML file** — `config.yml`
3. **Environment variables** — `XBANK_` prefix (e.g., `XBANK_APP_PORT=4000`)
4. **Legacy ENV** — `DATABASE_URL`, `REDIS_URL`, etc.

## Security

- PAN, PIN, CVV → AES-GCM encryption (E2EE)
- Passwords → bcrypt (cost=12)
- JWT → ES256 (ECDSA P-256)
- Request integrity → HMAC-SHA256
- CSRF protection on mutating endpoints
- Row-Level Security in PostgreSQL
- Rate limiting (per-IP and per-user)

## License

Private — All rights reserved.
