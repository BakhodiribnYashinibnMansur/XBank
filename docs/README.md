# XBank Documentation

## Arxitektura
- [Umumiy ko'rinish](architecture/overview.md) — Tech stack, bounded contexts, loyiha strukturasi
- [Event Sourcing](architecture/event-sourcing.md) — Account event sourcing, snapshots, CQRS
- [Saga Pattern](architecture/saga-pattern.md) — Transfer saga, compensations
- [Message Queue](architecture/message-queue.md) — Apache Kafka, Protobuf, DLQ, async processing

## Ishlash Flowlari
- [Umumiy tizim flow](architecture/system-flow.md) — Register, Login, Transfer to'liq diagramma
- [Backend flow](architecture/backend-flow.md) — Request lifecycle, middleware stack, error handling
- [Frontend flow](architecture/frontend-flow.md) — UI pages, SPA, SSE, auto-logout
- [Database flow](architecture/database-flow.md) — Transaction, hold, partitioning, backup, pg_cron

## Features
- [Auth & Session](features/auth-session.md) — JWT, 2FA/TOTP, session management, RBAC+ABAC
- [Account Management](features/accounts.md) — Account CRUD, hold mexanizmi, daily/monthly limits
- [Transfers](features/transfers.md) — Double-entry, idempotency, state machine, scheduled
- [Cards](features/cards.md) — PCI DSS, tokenizatsiya, Luhn, EMV
- [KYC & AML](features/kyc-aml.md) — Know Your Customer, Anti-Money Laundering, FATF
- [Fraud Detection](features/fraud-detection.md) — Risk scoring, velocity, device fingerprint
- [Beneficiaries](features/beneficiaries.md) — Transfer qiluvchilar ro'yxati
- [Notifications](features/notifications.md) — SSE real-time, audit trail
- [Exchange Rates](features/exchange-rates.md) — Valyuta kurslari, ISO 4217
- [Reconciliation](features/reconciliation.md) — Ledger integrity, daily checks

## Database
- [PostgreSQL](database/postgresql.md) — Schema, RLS, partitioning, pg_cron, backup
- [Transaction Layer](database/transactions.md) — ACID, isolation levels, locking, retry
<!-- TODO: database/migrations.md — Barcha migratsiyalar ro'yxati (hali yaratilmagan) -->

## Security
- [Security Overview](security/overview.md) — OWASP, encryption, hashing, rate limiting
- [Encryption & PKI](security/encryption.md) — E2EE (ECIES), ES256, Envelope encryption, key rotation, Vault
<!-- TODO: security/api-standards.md — Response format, pagination, Swagger (hali yaratilmagan) -->
- [Xalqaro Standartlar](security/compliance.md) — PCI DSS, ISO 27001, SOC 2, GDPR

## Test
<!-- TODO: features/test-ui.md — Frontend test interfeysi (hali yaratilmagan) -->
<!-- TODO: features/test-client.md — CLI E2E test (hali yaratilmagan) -->

## Implementatsiya
- [Faza ketma-ketligi](architecture/implementation-phases.md)
