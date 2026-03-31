# Implementation Phase Sequence

## Phase 1: Foundation + Git
1. `git init` + remote: `git@github.com:BakhodiribnYashinibnMansur/XBank.git`
2. `go mod init`, docker-compose (postgres, redis, pgbouncer, prometheus, grafana, loki)
3. `internal/domain/shared/` — AggregateRoot, EventSourced, Money (HALF_EVEN), ISO 4217
4. Makefile, Dockerfile, .env.example, .gitignore

## Phase 2: Identity + Session + 2FA + RBAC + Security
5. User, Session, TOTP, KYC domain + migrations
6. JWT (RS256), bcrypt, AES-256, HMAC, TOTP, bruteforce, IP whitelist, challenge
7. RBAC+ABAC middleware, device fingerprint
8. Fiber server + 14 middleware stack

## Phase 3: Account + Event Sourcing + CQRS
9. Event Sourced Account aggregate (balance, available_balance, hold)
10. Event Store + Snapshot Store (migrations)
11. CQRS read models / projections
12. DB: RLS, partitioning, isolation levels, pg_cron

## Phase 4: Transfer + Saga + AML + Fraud
13. Transfer Saga (10 steps, compensations)
14. AML screening + Fraud detection (risk scoring, velocity, device)
15. SERIALIZABLE isolation + pessimistic lock + HMAC signing
16. Idempotency + DLQ + scheduled transactions

## Phase 5: Card + Tokenization
17. Card aggregate (Luhn, EMV, 3DS)
18. AES encryption + tokenization + hold/capture/release

## Phase 6: Beneficiary + Exchange + Notifications + Queue
19. Beneficiary, Exchange, Audit, SSE, Kafka queue (Protobuf), Schema Registry
20. Reconciliation (daily)

## Phase 7: Monitoring
21. Prometheus metrics, Grafana dashboard, Loki, alerting

## Phase 8: Test UI + Test Client
22. Web UI (HTML/JS/Tailwind), CLI test client
23. Swagger docs, seed data
24. Integration + Security + Performance tests
