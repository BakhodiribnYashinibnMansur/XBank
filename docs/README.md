# XBank Documentation

## Architecture
- [Overview](architecture/overview.md) — Tech stack, bounded contexts, project structure
- [Event Sourcing](architecture/event-sourcing.md) — Account event sourcing, snapshots, CQRS
- [Saga Pattern](architecture/saga-pattern.md) — Transfer saga, compensations
- [Message Queue](architecture/message-queue.md) — Apache Kafka, Protobuf, DLQ, async processing

## Workflow Flows
- [System flow overview](architecture/system-flow.md) — Register, Login, Transfer full diagram
- [Backend flow](architecture/backend-flow.md) — Request lifecycle, middleware stack, error handling
- [Frontend flow](architecture/frontend-flow.md) — UI pages, SPA, SSE, auto-logout
- [Database flow](architecture/database-flow.md) — Transaction, hold, partitioning, backup, pg_cron

## Features
- [Auth & Session](features/auth-session.md) — JWT, 2FA/TOTP, session management, RBAC+ABAC
- [Account Management](features/accounts.md) — Account CRUD, hold mechanism, daily/monthly limits
- [Transfers](features/transfers.md) — Double-entry, idempotency, state machine, scheduled
- [Cards](features/cards.md) — PCI DSS, tokenization, Luhn, EMV
- [KYC & AML](features/kyc-aml.md) — Know Your Customer, Anti-Money Laundering, FATF
- [Fraud Detection](features/fraud-detection.md) — Risk scoring, velocity, device fingerprint
- [Beneficiaries](features/beneficiaries.md) — Transfer recipients list
- [Notifications](features/notifications.md) — SSE real-time, audit trail
- [Exchange Rates](features/exchange-rates.md) — Currency rates, ISO 4217
- [Reconciliation](features/reconciliation.md) — Ledger integrity, daily checks

## Database
- [PostgreSQL](database/postgresql.md) — Schema, RLS, partitioning, pg_cron, backup
- [Transaction Layer](database/transactions.md) — ACID, isolation levels, locking, retry
<!-- TODO: database/migrations.md — List of all migrations (not yet created) -->

## Security
- [Security Overview](security/overview.md) — OWASP, encryption, hashing, rate limiting
- [Encryption & PKI](security/encryption.md) — E2EE (ECIES), ES256, Envelope encryption, key rotation, Vault
<!-- TODO: security/api-standards.md — Response format, pagination, Swagger (not yet created) -->
- [International Standards](security/compliance.md) — PCI DSS, ISO 27001, SOC 2, GDPR

## Testing
<!-- TODO: features/test-ui.md — Frontend test interface (not yet created) -->
<!-- TODO: features/test-client.md — CLI E2E test (not yet created) -->

## Implementation
- [Phase sequence](architecture/implementation-phases.md)
