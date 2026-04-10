# Database Migrations

XBank uses [Goose](https://github.com/pressly/goose) for SQL migration management.

## Commands

```bash
make migrate-up                     # Apply all pending migrations
make migrate-down                   # Rollback the last migration
make migrate-status                 # View migration status
make migrate-create name=add_xyz    # Create a new migration
make migrate-reset                  # Rollback all migrations
```

## Migration Index

| # | File | Description | Tables |
|---|------|-------------|--------|
| 001 | `001_create_users.sql` | User accounts | `users` |
| 002 | `002_create_sessions.sql` | JWT sessions | `sessions` |
| 003 | `003_create_accounts.sql` | Bank accounts | `accounts` |
| 004 | `004_create_transfers.sql` | Money transfers | `transfers` |
| 005 | `005_create_cards.sql` | Payment cards | `cards` |
| 006 | `006_create_account_events.sql` | Account event sourcing | `account_events` |
| 007 | `007_create_transfer_events.sql` | Transfer event sourcing | `transfer_events` |
| 008 | `008_add_user_role.sql` | Add role column to users | ALTER `users` |
| 009 | `009_create_beneficiaries.sql` | Transfer recipients | `beneficiaries` |
| 010 | `010_create_exchange_rates.sql` | Currency exchange rates | `exchange_rates` |
| 011 | `011_create_kyc.sql` | KYC verification | `kyc_verifications` |
| 012 | `012_create_fraud_checks.sql` | Fraud detection | `fraud_checks` |
| 013 | `013_create_ledger.sql` | Double-entry ledger | `ledger_entries` |
| 014 | `014_create_device_fingerprints.sql` | Device tracking | `device_fingerprints` |
| 015 | `015_enable_rls.sql` | Row-Level Security policies | ALTER multiple tables |
| 016 | `016_create_admin_whitelist.sql` | Admin IP whitelist | `admin_whitelist_ips` |
| 017 | `017_create_snapshots.sql` | Event sourcing snapshots | `account_snapshots`, `transfer_snapshots` |
| 018 | `018_partition_event_stores.sql` | Partition event tables by month | `account_events_*`, `transfer_events_*` |
| 019 | `019_create_read_projections.sql` | CQRS read models | `account_projections`, `transfer_projections`, `daily_balance_projections` |
| 020 | `020_setup_pg_cron.sql` | Automated partition creation | pg_cron jobs |
| 021 | `021_create_challenges.sql` | Step-up authentication | `challenges` |
| 022 | `022_create_dead_letter_queue.sql` | Kafka DLQ storage | `dead_letter_queue` |
| 023 | `023_create_scheduled_transfers.sql` | Scheduled/recurring transfers | `scheduled_transfers` |
| 024 | `024_create_card_tokens.sql` | Card tokenization & holds | `card_tokens`, `card_holds` |
| 025 | `025_add_totp_to_users.sql` | TOTP 2FA columns | ALTER `users` |
| 026 | `026_create_outbox.sql` | Transactional outbox | `outbox` |
| 027 | `027_create_rbac_tables.sql` | Role-Based Access Control | `rbac_roles`, `rbac_permissions`, `rbac_policies` |
| 028 | `028_create_audit_tables.sql` | Audit logging (partitioned) | `audit_logs`, `endpoint_history` |
| 029 | `029_create_ddd_bc_tables.sql` | DDD BC support tables | `feature_flags`, `feature_flag_rule_groups`, `feature_flag_conditions`, `site_settings`, `data_exports`, `statistics_snapshots`, `notifications`, `translations`, `files`, `announcements`, `system_errors`, `error_codes` |
| 031 | `031_create_user_settings.sql` | User preferences | `user_settings` |
| 032 | `032_create_user_contacts.sql` | User contacts | `user_contacts` |
| 033 | `033_create_reconciliation_runs.sql` | Ledger reconciliation | `reconciliation_runs` |
| 034 | `034_fix_card_pan_column.sql` | Fix PAN column type | ALTER `cards` |
| 035 | `035_create_new_bc_tables.sql` | Ops/Admin BC tables | `rate_limit_rules`, `app_metrics`, `ip_rules`, `integrations` |
| 036 | `036_create_4_new_bc_tables.sql` | New BC tables | `currencies`, `templates`, `health_records` |

## Conventions

- Format: `NNN_description.sql` (NNN = sequence number)
- Goose annotations: `-- +goose Up` / `-- +goose Down` required
- Always include a **Down** migration for rollback
- Partitioned tables: parent + initial partitions in one file
- Use `CREATE TABLE IF NOT EXISTS` for idempotency
- Use `gen_random_uuid()` for UUID primary keys
- All timestamps: `TIMESTAMPTZ NOT NULL DEFAULT now()`

## Table Count

**Total: 60+ tables** across 36 migration files, including:
- 12 partitioned tables (event stores, audit logs, endpoint history)
- 3 CQRS read projection tables
- 3 RBAC tables
- 2 snapshot tables
- 1 outbox table
- 1 dead letter queue table
