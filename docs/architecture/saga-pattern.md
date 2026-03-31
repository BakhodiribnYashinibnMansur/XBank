# Saga Pattern — Transfer Orchestrator

## Transfer Saga (10 Steps + Compensations)

```
Step 1:  Fraud Check         → Compensate: -
Step 2:  AML Screening       → Compensate: -
Step 3:  Lock Source Account  → Compensate: Unlock
Step 4:  Validate Balance     → Compensate: Unlock
Step 5:  Lock Target Account  → Compensate: Unlock both
Step 6:  Debit Source         → Compensate: Refund Source
Step 7:  Credit Target        → Compensate: Reverse Credit + Refund
Step 8:  Create Ledger        → Compensate: Void entries
Step 9:  Complete Transfer    → Compensate: Mark Failed
Step 10: Emit Events          → -
```

## Saga State Machine

<!-- All states and transitions of the Transfer saga -->
```
                    ┌──────────┐
                    │  CREATED │
                    └────┬─────┘
                         │ start()
                         ▼
                    ┌──────────┐
            ┌──────│ CHECKING │ (Step 1-2: Fraud + AML)
            │      └────┬─────┘
            │           │ pass
            │           ▼
            │      ┌──────────┐
            │ ┌────│ LOCKING  │ (Step 3-5: Account locks)
            │ │    └────┬─────┘
            │ │         │ locked
            │ │         ▼
            │ │    ┌──────────┐
            │ │ ┌──│EXECUTING │ (Step 6-8: Debit, Credit, Ledger)
            │ │ │  └────┬─────┘
            │ │ │       │ done
            │ │ │       ▼
            │ │ │  ┌───────────┐
            │ │ │  │COMPLETING │ (Step 9-10: Complete + Events)
            │ │ │  └────┬──────┘
            │ │ │       │
            │ │ │       ▼
            │ │ │  ┌───────────┐
            │ │ │  │ COMPLETED │ ← final state
            │ │ │  └───────────┘
            │ │ │
            ▼ ▼ ▼       (if any step fails)
       ┌────────────┐
       │COMPENSATING│ → compensate in reverse order
       └─────┬──────┘
             │
             ▼
       ┌──────────┐
       │  FAILED  │ ← final state
       └──────────┘
```

## Saga State Persistence

<!-- Saga state is stored in DB — can resume even after server crash -->
```sql
CREATE TABLE saga_state (
    id              UUID PRIMARY KEY,
    transfer_id     UUID NOT NULL UNIQUE,
    current_step    INTEGER NOT NULL DEFAULT 0,    -- current step number
    status          VARCHAR(20) NOT NULL,          -- CREATED, CHECKING, LOCKING, EXECUTING, COMPLETING, COMPENSATING, COMPLETED, FAILED
    step_results    JSONB DEFAULT '{}',            -- result of each step
    error           TEXT,                          -- error message (if any)
    started_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    completed_at    TIMESTAMPTZ
);

-- For finding stuck sagas
CREATE INDEX idx_saga_active ON saga_state (status, updated_at)
    WHERE status NOT IN ('COMPLETED', 'FAILED');
```

### Saga Recovery (Server Crash)
```
When server restarts:
  1. Load sagas from saga_state where status != COMPLETED/FAILED
  2. For each saga:
     a. Resume from current_step (safe because it's idempotent)
     b. If timeout has passed → compensate
```

## Compensation

If Step 7 fails (in reverse order):
1. Step 6 compensate: Refund Source — return the debited amount
2. Step 5 compensate: Unlock both — release locks on both accounts
3. Step 4 compensate: (no action — was only a check)
4. Step 3 compensate: Unlock — release source account lock
5. Transfer status = FAILED

<!-- Each compensation must also be IDEMPOTENT —
     must produce the same result even if called multiple times -->

## Timeout and Retry

```
Saga Timeout:
  - Overall timeout: 30 seconds (for all steps)
  - Per-step timeout: 5 seconds
  - On timeout → compensation begins

Retry Strategy:
  - If step fails → 3 retries (100ms → 200ms → 400ms)
  - If all retries fail → compensate
  - Retryable errors: network error, DB connection, serialization failure
  - Not retryable: insufficient balance, account frozen, AML block
```

## Deadlock Prevention

Locking accounts in UUID order:
```go
// Always lock the smaller UUID first
// This ensures all concurrent transfers lock in the same order
if fromID > toID {
    lock(toID)    // smaller UUID first
    lock(fromID)  // larger UUID second
} else {
    lock(fromID)
    lock(toID)
}
```

## Monitoring and Alerting

<!-- The following metrics are tracked for monitoring sagas -->
```
Prometheus Metrics:
  - saga_duration_seconds        — saga execution time (histogram)
  - saga_step_duration_seconds   — time per step (histogram)
  - saga_total                   — total sagas (counter, label: status)
  - saga_compensation_total      — number of compensations (counter)
  - saga_active_count            — currently active sagas (gauge)

Alert Rules:
  - saga_active_count > 100          → "Too many concurrent sagas" alert
  - saga_duration_seconds > 30       → "Stuck saga" alert
  - saga_compensation_total rising fast → "Too many transfer failures" alert
```
