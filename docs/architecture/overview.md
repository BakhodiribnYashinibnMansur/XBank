# XBank — Arxitektura Umumiy Ko'rinish

## Tech Stack
- **Go 1.22+** + **GoFiber v2** (fasthttp)
- **pgx v5** + **PgBouncer** (transaction pool mode)
- **PostgreSQL 17** (write primary + read replica, partitioning, RLS)
- **Redis 7** (session, cache, rate limit, pub/sub)
- **Apache Kafka** (message queue, async processing) + **Protobuf** (message serialization)
- **Confluent Schema Registry** (Protobuf schema versioning, backward compatibility)
- **Prometheus + Grafana + Loki** (monitoring)
- **HashiCorp Vault** (key management — JWT, Card KEK, KYC KEK private keys)
- **Docker + docker-compose**

## Arxitektura: DDD Modular Monolith

Bounded contextlar Go package sifatida — microservice'ga o'tish oson.

| Context | Aggregate Roots | Mas'uliyat |
|---|---|---|
| **Identity** | `User`, `Session` | Auth, 2FA, KYC, RBAC |
| **Account** | `Account` | Event Sourced, double-entry ledger, hold |
| **Transfer** | `Transfer` | Saga, AML, idempotency, ECDSA signing |
| **Card** | `Card` | PCI DSS, tokenizatsiya |
| **Beneficiary** | `Beneficiary` | Transfer qiluvchilar |
| **Exchange** | `ExchangeRate` | Valyuta kurslari |
| **Notification** | `AuditRecord` | SSE, alerts, audit |
| **Fraud** | `FraudCheck` | Risk scoring, velocity |
| **Crypto** | `EncryptionKey`, `SigningKey` | PKI, key rotation, Vault integration |

## Asosiy Patternlar

- **Event Sourcing** — Account holati eventlar yig'indisi
- **CQRS** — Write (primary) / Read (replica) ajratilgan
- **Saga** — Transfer multi-step orchestrator
- **Double-Entry Bookkeeping** — Har bir transfer = debit + credit
- **Specification Pattern** — Business rules (SufficientBalance, DailyLimit)
- **Unit of Work** — Transactional boundary = aggregate boundary

## Loyiha Strukturasi

```
XBank/
├── cmd/api/main.go                    # Entry point
├── cmd/testclient/main.go             # E2E test client
├── internal/
│   ├── domain/                        # Pure domain — ZERO external import
│   │   ├── shared/                    # AggregateRoot, Money, Currency, Event, Spec
│   │   ├── identity/                  # User, Session, TOTP, KYC
│   │   ├── account/                   # Event Sourced Account, Ledger, Snapshot
│   │   ├── transfer/                  # Transfer Saga, AML
│   │   ├── card/                      # Card, Token
│   │   ├── beneficiary/               # Beneficiary
│   │   ├── fraud/                     # FraudCheck, Velocity, Device
│   │   └── exchange/                  # ExchangeRate
│   ├── application/                   # CQRS Command + Query handlers
│   │   └── crypto/                    # EncryptionKey, SigningKey, KeyRotation
│   ├── infrastructure/                # DB, Redis, Kafka, Auth, Crypto, Vault
│   └── interfaces/http/              # Fiber handlers, middleware, DTO
├── web/static/                        # Test UI
├── docs/                              # Documentation
├── deployments/                       # Infra configs
└── docker-compose.yml
```

## Middleware Stack (14 ta)

```
1.  Recovery           — Panic recovery
2.  RequestID          — Unique request ID
3.  Correlation-ID     — X-Correlation-ID tracking
4.  Helmet             — Security headers (CSP, HSTS)
5.  CORS               — Cross-Origin
6.  CSRF               — CSRF token
7.  Rate Limit         — Sliding window (Redis)
8.  Prometheus Metrics  — Request metrics
9.  Audit Logger       — Request/Response logging
10. Session            — JWT validation
11. RBAC/ABAC          — Authorization
12. 2FA                — Two-factor enforcement
13. KYC Required       — KYC status check
14. Device Fingerprint — Device tracking
```
