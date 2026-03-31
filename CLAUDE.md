# XBank — Claude Code Yo'riqnomasi

## Loyiha Haqida
XBank — Go (Fiber) + PostgreSQL asosida qurilgan banking application.
DDD Modular Monolith arxitektura, Event Sourcing, CQRS, Saga pattern.

## Tech Stack
- **Go 1.25+** (GoFiber v2, pgx v5)
- **PostgreSQL 17** (primary + read replica)
- **Redis 7** (session, cache, rate limit, pub/sub)
- **Apache Kafka** (async message queue, Protobuf serialization)
- **HashiCorp Vault** (key management, E2EE)
- **Docker + docker-compose**

## Loyiha Strukturasi
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

## Komandalar

### Ishga tushirish
```bash
make run              # Lokal server
make docker-up        # Docker bilan
make docker-down      # Docker to'xtatish
make build            # Binary build
make test             # Testlar
```

### Migratsiyalar (Goose)
```bash
make migrate-up                     # Barcha migratsiyalarni qo'llash
make migrate-down                   # Oxirgi migratsiyani qaytarish
make migrate-status                 # Holat ko'rish
make migrate-create name=add_cards  # Yangi migratsiya yaratish
make migrate-reset                  # Barcha migratsiyalarni qaytarish
```

### Goose o'rnatish
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

## Arxitektura Qoidalari

### Domain Layer (`internal/domain/`)
- **ZERO** external import (stdlib only)
- Aggregate Root = tranzaksiya chegarasi
- Domain event lar orqali holat o'zgarishi
- Specification pattern → business rules

### Application Layer (`internal/application/`)
- CQRS: Command (write) va Query (read) ajratilgan
- Command → primary DB (SERIALIZABLE)
- Query → read replica

### Infrastructure Layer (`internal/infrastructure/`)
- Repository implementation (pgx)
- Redis, Kafka, Vault integratsiya
- JWT, Crypto

### HTTP Layer (`internal/interfaces/http/`)
- Fiber handlers
- 14 ta middleware stack
- DTO validation

## Migration Qoidalari
- Goose format: `-- +goose Up` / `-- +goose Down` annotatsiyalar
- Har doim **Down** migratsiya yozing (rollback uchun)
- Fayl nomi: `NNN_description.sql` (NNN = tartib raqami)
- Partitioned jadvallar uchun: parent + dastlabki partitsiya bitta faylda

## Security Qoidalari
- PAN, PIN, CVV → **E2EE (ECIES)**: server plaintext ko'rmasin
- Parol, PIN, CVV → **bcrypt** (hash, cost=12)
- JWT → **ES256** (ECDSA P-256)
- Sensitive data logga **HECH QACHON** plaintext yozilmasin
- E2EE encrypted field lar logda `[E2EE_REDACTED]`
- Authorization header logda `Bearer [REDACTED]`

## Kafka Message Qoidalari
- Message format: **Protobuf** (JSON emas)
- Topic naming: `xbank.{domain}.{event}` (dot separator)
- Partition key: `account_id` yoki `user_id` (ordering uchun)
- Proto fayllar: `proto/` papkada

## Kodlash Standartlari
- Go standart formatting (`gofmt`)
- Error handling: wrap with context (`fmt.Errorf("... : %w", err)`)
- Naming: Go conventions (exported = PascalCase, unexported = camelCase)
- Commit message: `type: description` (docs:, feat:, fix:, refactor:, test:)
