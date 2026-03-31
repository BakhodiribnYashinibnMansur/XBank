# Account Management

## Event Sourced Account

Account state is the sum of events. Never UPDATE/DELETE — only APPEND.

### Account Aggregate
```go
type Account struct {
    EventSourcedAggregate
    UserID           uuid.UUID
    Number           AccountNumber
    Currency         Currency
    Status           AccountStatus      // ACTIVE, FROZEN, CLOSED
    Balance          Money              // Total balance
    AvailableBalance Money              // Available to use (balance - holds)
    HoldAmount       Money              // Blocked amount
    DailyDebitLimit  Money
    MonthlyDebitLimit Money
}
```

**Invariant:** `Balance = AvailableBalance + HoldAmount`

### Operations
```go
func (a *Account) Open(userID, currency)           // → AccountOpenedEvent
func (a *Account) Credit(amount, ref)               // → AccountCreditedEvent
func (a *Account) Debit(amount, ref)                // → AccountDebitedEvent (SufficientBalanceSpec)
func (a *Account) PlaceHold(amount, ref)            // → HoldPlacedEvent
func (a *Account) CaptureHold(ref, captureAmount)   // → HoldCapturedEvent (partial capture)
func (a *Account) ReleaseHold(ref)                  // → HoldReleasedEvent
func (a *Account) Freeze()                          // → AccountFrozenEvent
    // On a frozen account: new debit/credit BLOCKED, existing holds are kept until they expire
    // Frozen by admin or reconciliation system
func (a *Account) Unfreeze()                        // → AccountUnfrozenEvent
func (a *Account) Close()                           // → AccountClosedEvent
    // Close rules:
    //   1. balance MUST be 0
    //   2. hold_amount MUST be 0 (cannot close if there are existing holds)
    //   3. There must be no pending transfers
```

## Hold Mechanism

```
Placing a hold (card authorization):
  available_balance -= amount
  hold_amount += amount

Hold capture (payment confirmation):
  balance -= amount (or partial amount)
  hold_amount -= amount

Hold release (cancellation):
  available_balance += amount
  hold_amount -= amount
```

## Daily/Monthly Limits

```sql
-- Account level:
daily_debit_limit   BIGINT DEFAULT 1000000000   -- $10K
monthly_debit_limit BIGINT DEFAULT 10000000000  -- $100K

-- User level:
daily_transfer_limit  BIGINT DEFAULT 500000000   -- $5K
monthly_transfer_limit BIGINT DEFAULT 5000000000  -- $50K
```

Check before transfer:
1. `SUM(debits today) + amount <= daily_debit_limit`
2. `SUM(all account debits today) + amount <= daily_transfer_limit`
3. Monthly works the same way

## Specifications
- `SufficientBalanceSpec` — available_balance >= amount
- `DailyLimitSpec` — within daily limit
- `MonthlyLimitSpec` — within monthly limit
- `AccountActiveSpec` — status == ACTIVE

## Event Store + Snapshots

Save snapshot after every 100 events:
```go
// Loading: snapshot + events after it
snapshot := loadLatestSnapshot(id)     // version=100
events := loadEventsAfter(id, 100)    // 101, 102, ...
account := rebuildFromSnapshot(snapshot)
account.Replay(events)
```

## CQRS Read Models
- `AccountSummaryView` — balance, last tx, account info
- `TransactionHistoryView` — paginated, filtered
- `DashboardView` — all accounts + total balance

## API Endpoints

| Method | Endpoint | Middleware |
|---|---|---|
| POST | `/api/v1/accounts` | Session+KYC |
| GET | `/api/v1/accounts` | Session |
| GET | `/api/v1/accounts/{id}` | Session |
| POST | `/api/v1/accounts/{id}/close` | Session+2FA |
| GET | `/api/v1/accounts/{id}/statements` | Session |
