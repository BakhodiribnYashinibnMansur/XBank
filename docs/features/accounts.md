# Account Management

## Event Sourced Account

Account holati eventlar yig'indisi. Hech qachon UPDATE/DELETE — faqat APPEND.

### Account Aggregate
```go
type Account struct {
    EventSourcedAggregate
    UserID           uuid.UUID
    Number           AccountNumber
    Currency         Currency
    Status           AccountStatus      // ACTIVE, FROZEN, CLOSED
    Balance          Money              // Umumiy balans
    AvailableBalance Money              // Foydalanish mumkin (balance - holds)
    HoldAmount       Money              // Bloklangan summa
    DailyDebitLimit  Money
    MonthlyDebitLimit Money
}
```

**Invariant:** `Balance = AvailableBalance + HoldAmount`

### Operatsiyalar
```go
func (a *Account) Open(userID, currency)           // → AccountOpenedEvent
func (a *Account) Credit(amount, ref)               // → AccountCreditedEvent
func (a *Account) Debit(amount, ref)                // → AccountDebitedEvent (SufficientBalanceSpec)
func (a *Account) PlaceHold(amount, ref)            // → HoldPlacedEvent
func (a *Account) CaptureHold(ref, captureAmount)   // → HoldCapturedEvent (partial capture)
func (a *Account) ReleaseHold(ref)                  // → HoldReleasedEvent
func (a *Account) Freeze()                          // → AccountFrozenEvent
    // Frozen account da: yangi debit/credit BLOCK, mavjud hold lar expire bo'lguncha saqlanadi
    // Admin yoki reconciliation tizimi muzlatadi
func (a *Account) Unfreeze()                        // → AccountUnfrozenEvent
func (a *Account) Close()                           // → AccountClosedEvent
    // Close qoidalari:
    //   1. balance = 0 bo'lishi SHART
    //   2. hold_amount = 0 bo'lishi SHART (mavjud hold lar bo'lsa close mumkin emas)
    //   3. Pending transfer lar bo'lmasligi kerak
```

## Hold Mexanizmi

```
Hold qo'yish (karta authorization):
  available_balance -= amount
  hold_amount += amount

Hold capture (to'lov tasdiqlash):
  balance -= amount (yoki partial amount)
  hold_amount -= amount

Hold release (bekor qilish):
  available_balance += amount
  hold_amount -= amount
```

## Daily/Monthly Limits

```sql
-- Account darajasida:
daily_debit_limit   BIGINT DEFAULT 1000000000   -- $10K
monthly_debit_limit BIGINT DEFAULT 10000000000  -- $100K

-- User darajasida:
daily_transfer_limit  BIGINT DEFAULT 500000000   -- $5K
monthly_transfer_limit BIGINT DEFAULT 5000000000  -- $50K
```

Transfer oldidan tekshirish:
1. `SUM(debits today) + amount <= daily_debit_limit`
2. `SUM(all account debits today) + amount <= daily_transfer_limit`
3. Monthly ham xuddi shunday

## Specifications
- `SufficientBalanceSpec` — available_balance >= amount
- `DailyLimitSpec` — kunlik limit ichida
- `MonthlyLimitSpec` — oylik limit ichida
- `AccountActiveSpec` — status == ACTIVE

## Event Store + Snapshots

Har 100 eventdan keyin snapshot saqlash:
```go
// Yuklash: snapshot + undan keyingi eventlar
snapshot := loadLatestSnapshot(id)     // version=100
events := loadEventsAfter(id, 100)    // 101, 102, ...
account := rebuildFromSnapshot(snapshot)
account.Replay(events)
```

## CQRS Read Models
- `AccountSummaryView` — balance, last tx, account info
- `TransactionHistoryView` — paginated, filtered
- `DashboardView` — barcha accountlar + umumiy balance

## API Endpoints

| Method | Endpoint | Middleware |
|---|---|---|
| POST | `/api/v1/accounts` | Session+KYC |
| GET | `/api/v1/accounts` | Session |
| GET | `/api/v1/accounts/{id}` | Session |
| POST | `/api/v1/accounts/{id}/close` | Session+2FA |
| GET | `/api/v1/accounts/{id}/statements` | Session |
