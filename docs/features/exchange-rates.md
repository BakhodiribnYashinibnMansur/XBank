# Exchange Rates — ISO 4217

## Currency Value Object
```go
type Currency struct {
    Code     string // "UZS", "USD", "EUR"
    Exponent int    // UZS=2, USD=2, EUR=2
}
```

<!-- According to the ISO 4217 standard, the number of
     minor units (tiyin, cent) is defined for each currency -->

## Money Value Object (Immutable)
```go
type Money struct {
    amount   *big.Int  // minor units (tiyin/cent)
    currency Currency
}
// NEVER float!
// 100,000.50 so'm = 10000050 (tiyin)
// $99.99 = 9999 (cent)
// Banker's Rounding (HALF_EVEN)
```

### Banker's Rounding Examples
<!-- HALF_EVEN — when the value is 0.5, round to the nearest even number.
     This is the standard in financial calculations because it reduces bias. -->
```
Standard rounding:             Banker's Rounding (HALF_EVEN):
  2.5 → 3                       2.5 → 2  (2 is even, round down)
  3.5 → 4                       3.5 → 4  (4 is even, round up)
  4.5 → 5                       4.5 → 4  (4 is even, round down)
  5.5 → 6                       5.5 → 6  (6 is even, round up)
  2.3 → 2                       2.3 → 2  (normal, no difference)
  2.7 → 3                       2.7 → 3  (normal, no difference)

Practical example (currency conversion):
  $100 * rate 12,750.5 UZS = 1,275,050 tiyin
  If there is a 0.5 tiyin remainder → round to the nearest even number
```

## Exchange Rate Aggregate
```go
type ExchangeRate struct {
    AggregateRoot
    Pair      CurrencyPair  // USD/UZS
    Rate      RateValue     // rate * 1000000 (6 decimal precision)
    Spread    RateValue     // bank spread/markup (buy-sell difference)
    BuyRate   RateValue     // buy rate (bank sells)
    SellRate  RateValue     // sell rate (bank buys)
    ValidFrom time.Time     // rate start time
    ValidTo   time.Time     // rate validity period
    Source    string        // rate source ("CBU", "MANUAL", "API")
}
```

## Rate Source

<!-- Rates are obtained from the following sources -->
```
1. CBU (Central Bank) API — official rate, updated at the start of the day
2. Manual entry — setting a custom rate via the admin panel
3. External API — real-time rates (in the future)

Update schedule:
  - CBU rate is loaded automatically every day at 09:00
  - Admin can set a manual rate at any time
  - Manual rate takes priority over CBU rate
```

## Rate Locking

<!-- The rate may change during a transfer.
     Therefore, the rate is locked when the transfer begins. -->
```
Transfer Flow:
  1. User initiates a transfer (USD → UZS)
  2. Server retrieves the current rate and locks it for 2 minutes
  3. The locked rate is displayed to the user
  4. User confirms → transfer is executed at this rate
  5. If not confirmed within 2 minutes → rate lock expires

Rate Lock table:
  rate_lock_id    UUID
  pair            VARCHAR(7)     -- "USD/UZS"
  locked_rate     BIGINT         -- locked rate
  user_id         UUID
  expires_at      TIMESTAMPTZ    -- after 2 minutes
  used            BOOLEAN        -- whether used in a transfer
```

## FX Spread (Buy-Sell Difference)

```
Example:
  CBU official rate: 1 USD = 12,750 UZS
  Bank buy rate:     1 USD = 12,700 UZS  (bank buys)
  Bank sell rate:    1 USD = 12,800 UZS  (bank sells)
  Spread:            100 UZS (0.78%)

When the user makes a transfer:
  - USD → UZS: sell rate is used (12,800)
  - UZS → USD: buy rate is used (12,700)
```

## Stale Rate Handling

<!-- If the cached rate is stale -->
```
Rules:
  - Redis cache: 5 min TTL (for regular viewing)
  - For transfers: ALWAYS get a fresh rate from the DB (not cache!)
  - If the DB rate is older than 24 hours:
    → Transfer is BLOCKed
    → Admin alert is sent
    → "Rate not updated" error message
```

## API

| Method | Endpoint | Middleware | Description |
|---|---|---|---|
| GET | `/api/v1/currencies` | Public | Supported currencies |
| GET | `/api/v1/currencies/rates?from=USD&to=UZS` | Public | Current rate (cached) |
| POST | `/api/v1/currencies/rate-lock` | Session | Lock rate for transfer |

Redis cache: 5 min TTL (only for GET /rates, not for transfers).
