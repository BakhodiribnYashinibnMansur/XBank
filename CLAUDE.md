# XBank — Claude Code Guide

## About the Project
XBank — a banking application built on Go (Fiber) + PostgreSQL.
DDD Modular Monolith architecture, Event Sourcing, CQRS, Saga pattern.

## Tech Stack
- **Go 1.25+** (GoFiber v2, pgx v5)
- **PostgreSQL 17** (primary + read replica)
- **Redis 7** (session, cache, rate limit, pub/sub)
- **Apache Kafka** (async message queue, Protobuf serialization)
- **HashiCorp Vault** (key management, E2EE)
- **Docker + docker-compose**

## Project Structure
```
cmd/api/           → Entry point
internal/
  domain/          → Pure domain (ZERO external import)
  application/     → CQRS Command/Query handlers
  infrastructure/  → DB, Redis, Kafka, Auth, Crypto
  interfaces/http/ → Fiber handlers, middleware, DTO
migrations/        → Goose SQL migrations
docs/              → Documentation
proto/             → Protobuf definitions (Kafka messages)
```

## Commands

### Running
```bash
make run              # Local server
make docker-up        # With Docker
make docker-down      # Stop Docker
make build            # Binary build
make test             # Tests
```

### Migrations (Goose)
```bash
make migrate-up                     # Apply all migrations
make migrate-down                   # Rollback the last migration
make migrate-status                 # View status
make migrate-create name=add_cards  # Create a new migration
make migrate-reset                  # Rollback all migrations
```

### Installing Goose
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

## Architecture Rules

### Domain Layer (`internal/domain/`)
- **ZERO** external import (stdlib only)
- Aggregate Root = transaction boundary
- State changes through domain events
- Specification pattern → business rules

### Application Layer (`internal/application/`)
- CQRS: Command (write) and Query (read) separated
- Command → primary DB (SERIALIZABLE)
- Query → read replica

### Infrastructure Layer (`internal/infrastructure/`)
- Repository implementation (pgx)
- Redis, Kafka, Vault integration
- JWT, Crypto

### HTTP Layer (`internal/interfaces/http/`)
- Fiber handlers
- 14-middleware stack
- DTO validation

## Migration Rules
- Goose format: `-- +goose Up` / `-- +goose Down` annotations
- Always write a **Down** migration (for rollback)
- File name: `NNN_description.sql` (NNN = sequence number)
- For partitioned tables: parent + initial partition in one file

## Security Rules
- PAN, PIN, CVV → **E2EE (ECIES)**: server must not see plaintext
- Password, PIN, CVV → **bcrypt** (hash, cost=12)
- JWT → **ES256** (ECDSA P-256)
- Sensitive data must **NEVER** be written to logs as plaintext
- E2EE encrypted fields in logs: `[E2EE_REDACTED]`
- Authorization header in logs: `Bearer [REDACTED]`

## Kafka Message Rules
- Message format: **Protobuf** (not JSON)
- Topic naming: `xbank.{domain}.{event}` (dot separator)
- Partition key: `account_id` or `user_id` (for ordering)
- Proto files: in `proto/` directory

## Coding Standards
- Go standard formatting (`gofmt`)
- Error handling: wrap with context (`fmt.Errorf("... : %w", err)`)
- Naming: Go conventions (exported = PascalCase, unexported = camelCase)
- Commit message: `type: description` (docs:, feat:, fix:, refactor:, test:)
