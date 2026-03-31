# Reconciliation — Daily Verification

<!-- Reconciliation — the process of verifying the accuracy of financial data.
     One of the most important processes in a banking system.
     Automatic and real-time checks are performed daily. -->

## Daily Checks (every day at 3:00 AM, pg_cron)

### 1. Double-Entry Integrity
<!-- The sum of all ledger entries must ALWAYS equal 0.
     If not 0 → money has been "lost" or "extra was created". -->
```sql
SELECT CASE WHEN SUM(
    CASE WHEN side='DEBIT' THEN -amount_minor ELSE amount_minor END
) = 0 THEN 'OK' ELSE 'MISMATCH!' END
FROM ledger_entries
WHERE created_at >= CURRENT_DATE - INTERVAL '1 day';
-- Check for the previous day
-- MUST be 0 — otherwise a serious issue!
```

### 2. Account Balance = Ledger Sum
<!-- Each account balance must equal the sum of all its ledger entries -->
```sql
SELECT a.id, a.balance_minor,
       COALESCE(SUM(
           CASE WHEN le.side='CREDIT' THEN le.amount_minor
                ELSE -le.amount_minor END
       ), 0) AS calculated_balance
FROM accounts a
LEFT JOIN ledger_entries le ON le.account_id = a.id
GROUP BY a.id, a.balance_minor
HAVING a.balance_minor != COALESCE(SUM(
    CASE WHEN le.side='CREDIT' THEN le.amount_minor
         ELSE -le.amount_minor END
), 0);
-- Empty result = OK (all balances are correct)
-- Row exists = ALERT! (balance and ledger do not match)
```

### 3. Event Store Integrity
<!-- In event sourcing, account state = the result of replaying all events -->
```
For each account:
  1. Get Account.balance (value in DB)
  2. Replay all events from the beginning → calculated balance
  3. DB balance == calculated balance? → OK
  4. Not equal? → ALERT! + account freeze
```

### 4. Snapshot Validation
<!-- Verify snapshot correctness — the snapshot's state value
     must equal the result of replaying events up to that version -->
```
For each snapshot:
  1. Load snapshot (version=N, state=S1)
  2. Replay events from 1 to N → state S2
  3. S1 == S2? → OK
  4. S1 != S2? → Snapshot is corrupted → recreate + ALERT
```

### 5. Transfer Status Consistency
<!-- All COMPLETED transfers must have ledger entries -->
```sql
SELECT t.id FROM transfers t
WHERE t.status = 'COMPLETED'
AND NOT EXISTS (
    SELECT 1 FROM ledger_entries le WHERE le.transfer_id = t.id
);
-- COMPLETED transfer but no ledger entry = SERIOUS ISSUE
```

## Real-time Checks

<!-- In addition to daily checks, some checks are performed in real-time -->
```
After each transfer (inline):
  1. Source account: balance >= 0 (via CHECK constraint)
  2. Ledger: debit + credit = 0 (application level)
  3. Event version: monotonically increasing (UNIQUE constraint)

Every 1 hour (pg_cron):
  1. PENDING transfers > 5 minutes → ALERT (stuck transfer)
  2. Hold amount > 24 hours → ALERT (expired hold)
```

## Mismatch Handling

```
Severity levels:

CRITICAL (automatic action):
  - Double-entry integrity fail → all transfers are HALTED
  - Audit log + admin alert + SMS
  - System switches to read-only mode until manual review

HIGH (account level):
  - Account balance != ledger sum → account FREEZE
  - Account balance != event replay → account FREEZE
  - Admin sends for manual review

MEDIUM (data quality):
  - Snapshot invalid → snapshot recreation (automatic)
  - Transfer status inconsistency → sent to admin review queue

Resolution process:
  1. ALERT notification (admin + oncall)
  2. Audit log entry (RECONCILIATION_MISMATCH, with details)
  3. FREEZE relevant account(s) (if serious)
  4. Admin review:
     a. Identify root cause (review event log, audit trail)
     b. Fix (manual adjustment or bug fix)
     c. UNFREEZE account
  5. Write post-mortem
```

## Reconciliation Results Log

```sql
CREATE TABLE reconciliation_runs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_type        VARCHAR(20) NOT NULL,     -- DAILY, HOURLY, MANUAL
    started_at      TIMESTAMPTZ NOT NULL,
    completed_at    TIMESTAMPTZ,
    status          VARCHAR(20) NOT NULL,     -- RUNNING, PASSED, FAILED
    checks_passed   INTEGER DEFAULT 0,        -- how many checks passed
    checks_failed   INTEGER DEFAULT 0,        -- how many checks failed
    details         JSONB,                    -- result of each check
    accounts_frozen INTEGER DEFAULT 0,        -- how many accounts were frozen
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_recon_runs ON reconciliation_runs (run_type, started_at DESC);
```

## Performance

<!-- In a large system, reconciliation may be slow -->
```
Optimization:
  - Daily check only for the previous day (fast due to partitioned tables)
  - Event replay — start from snapshot (not 100 events together, just 1-2 events)
  - Parallel checking — each account in a separate goroutine
  - Weekly full check — for all history (on weekends, low load)

Estimated time:
  - 10,000 accounts: ~30 seconds
  - 100,000 accounts: ~5 minutes
  - 1,000,000 accounts: ~1 hour (parallel, with snapshots)
```
