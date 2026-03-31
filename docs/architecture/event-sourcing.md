# Event Sourcing

## Concept
```
Traditional approach:          Event Sourcing:
┌─────────┐               ┌─────────────────────────┐
│ balance  │               │ Event 1: +1000 (deposit) │
│  = 500   │               │ Event 2: -200 (transfer) │
└─────────┘               │ Event 3: -300 (payment)  │
                          │ Current state = 500       │
                          └─────────────────────────┘
```

- Every change is stored as an event
- Never UPDATE/DELETE — only APPEND
- You can "go back" and view the state at any point in time
- Ideal for audit — required by regulators

## Event Types

<!-- All Account aggregate event types -->

| Event | Description | Data (event_data) |
|---|---|---|
| `AccountOpenedEvent` | New account opened | `{user_id, currency, account_number}` |
| `AccountCreditedEvent` | Funds deposited to account | `{amount, reference, source}` |
| `AccountDebitedEvent` | Funds withdrawn from account | `{amount, reference, destination}` |
| `HoldPlacedEvent` | Funds held (card auth) | `{amount, reference, expires_at}` |
| `HoldCapturedEvent` | Hold confirmed (payment) | `{reference, captured_amount, original_amount}` |
| `HoldReleasedEvent` | Hold cancelled | `{reference, released_amount}` |
| `AccountFrozenEvent` | Account frozen | `{reason, frozen_by}` |
| `AccountUnfrozenEvent` | Account reactivated | `{reason, unfrozen_by}` |
| `AccountClosedEvent` | Account closed | `{reason, closed_by}` |

### Event Structure (Go)
```go
// Every event implements the following interface
type DomainEvent interface {
    EventType() string      // "AccountCreditedEvent"
    AggregateID() uuid.UUID // account UUID
    Version() int           // monotonically increasing number
    OccurredAt() time.Time
}
```

## Event Store
```sql
CREATE TABLE event_store (
    id             UUID PRIMARY KEY,
    aggregate_id   UUID NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL,       -- 'Account'
    event_type     VARCHAR(100) NOT NULL,      -- 'AccountCreditedEvent'
    event_data     JSONB NOT NULL,             -- event payload
    metadata       JSONB,                      -- correlation_id, user_id, ip, device_id
    version        INTEGER NOT NULL,           -- aggregate version (1, 2, 3, ...)
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(aggregate_id, version)              -- version cannot repeat for a single aggregate
) PARTITION BY RANGE (created_at);

-- Indexes
CREATE INDEX idx_event_store_aggregate ON event_store (aggregate_id, version);
CREATE INDEX idx_event_store_type ON event_store (event_type, created_at);
```

## Event Versioning (Schema Evolution)

<!-- Event structure may change over time.
     For example, a new field may need to be added to AccountCreditedEvent.
     Strategy for adding new fields without breaking old events: -->

```
Strategy: Upcasting (converting old format to new format at read time)

v1: { "amount": 100000 }
v2: { "amount": 100000, "source": "TRANSFER" }    ← new field added

Rules:
1. Adding a new field to event data — OK (with default value)
2. Removing an existing field — NOT ALLOWED
3. Renaming a field — NOT ALLOWED
4. Creating a new event type — OK (old event type is preserved)

In Go:
  func (e *AccountCreditedEvent) Upcast(data map[string]any) {
      if _, ok := data["source"]; !ok {
          data["source"] = "UNKNOWN"  // default for old events
      }
  }
```

## Snapshot (Performance)

Cache the current state after every 100 events:
```sql
CREATE TABLE snapshots (
    aggregate_id   UUID NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL,
    version        INTEGER NOT NULL,           -- events up to this version
    state          JSONB NOT NULL,             -- current state of the aggregate
    checksum       VARCHAR(64),                -- SHA-256 (integrity check)
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (aggregate_id, version)
);
```

### Snapshot Validation
<!-- Checking snapshot correctness during reconciliation -->
```
1. Load snapshot (version=N, state=S1)
2. Replay all events from 1 to N → state S2
3. S1 == S2? → OK
4. S1 != S2? → Snapshot corrupted, recreate + ALERT
   - Corrupted snapshot is deleted
   - Rebuilt from all events
   - New snapshot is saved
```

## Account Loading Strategy
```
AccountRepository.FindByID(id):
  │
  ├── 1. Get latest snapshot (version=100)
  │       → if no snapshot found, start from version=0
  │
  ├── 2. Get subsequent events (101, 102, ...)
  │       → if no events found, snapshot state = current state
  │
  ├── 3. Snapshot + replay = current state
  │
  └── 4. If new version % 100 == 0 → save new snapshot
```

## Temporal Query (Viewing state at a point in time)

<!-- For regulator or audit: "What was the balance on 2026-01-15?" -->
```
AccountRepository.FindByIDAtTime(id, targetTime):
  1. Get the latest snapshot before targetTime
  2. Replay events from snapshot to targetTime
  3. Result = exact state at that time

Use cases:
  - Audit query: "Balance on this date?"
  - Dispute: "State at the time of transfer?"
  - Regulator report: "Balance at end of month?"
```

## CQRS Projections

Event store → denormalized read model (materialized views):

| Projection | Purpose | Updated |
|---|---|---|
| `AccountSummaryView` | Account state: balance, status, last tx | On every account event |
| `TransactionHistoryView` | Paginated transaction history | On Credit/Debit/Transfer event |
| `DashboardView` | All accounts + total balance | On every balance change |

### Projection Update Mechanism
```
Event Store INSERT (after commit)
    │
    ├── Sync: EventBus.Publish()
    │       └── ProjectionHandler.Handle(event)
    │               ├── AccountSummaryProjection.Update(event)
    │               ├── TransactionHistoryProjection.Update(event)
    │               └── DashboardProjection.Update(event)
    │
    └── Async: Redis Streams (background)
            └── Consumer: ProjectionRebuilder (if projection is corrupted)
```

### Projection Rebuild
<!-- If projection is incorrect or a new projection is added -->
```
1. Create new projection table
2. Read ALL events sequentially from event_store
3. Pass each event to the projection handler
4. Replace old projection (atomic swap)
```
