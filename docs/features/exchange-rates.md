# Exchange Rates — ISO 4217

## Currency Value Object
```go
type Currency struct {
    Code     string // "UZS", "USD", "EUR"
    Exponent int    // UZS=2, USD=2, EUR=2
}
```

<!-- ISO 4217 standartiga muvofiq har bir valyuta uchun
     minor unit (tiyin, cent) soni belgilanadi -->

## Money Value Object (Immutable)
```go
type Money struct {
    amount   *big.Int  // minor units (tiyin/cent)
    currency Currency
}
// HECH QACHON float!
// 100,000.50 so'm = 10000050 (tiyin)
// $99.99 = 9999 (cent)
// Banker's Rounding (HALF_EVEN)
```

### Banker's Rounding Misollari
<!-- HALF_EVEN — 0.5 bo'lganda eng yaqin juft songa yaxlitlash.
     Bu moliyaviy hisob-kitoblarda standart, chunki bias (og'ish) kamaytiradi. -->
```
Oddiy yaxlitlash:           Banker's Rounding (HALF_EVEN):
  2.5 → 3                     2.5 → 2  (2 juft, pastga)
  3.5 → 4                     3.5 → 4  (4 juft, yuqoriga)
  4.5 → 5                     4.5 → 4  (4 juft, pastga)
  5.5 → 6                     5.5 → 6  (6 juft, yuqoriga)
  2.3 → 2                     2.3 → 2  (oddiy, farq yo'q)
  2.7 → 3                     2.7 → 3  (oddiy, farq yo'q)

Amaliy misol (valyuta konvertatsiya):
  $100 * kurs 12,750.5 UZS = 1,275,050 tiyin
  Agar 0.5 tiyin qoldiq bo'lsa → juft songa yaxlitlash
```

## Exchange Rate Aggregate
```go
type ExchangeRate struct {
    AggregateRoot
    Pair      CurrencyPair  // USD/UZS
    Rate      RateValue     // rate * 1000000 (6 decimal precision)
    Spread    RateValue     // bank spread/markup (oldi-sotdi farqi)
    BuyRate   RateValue     // sotib olish kursi (bank sotadi)
    SellRate  RateValue     // sotish kursi (bank sotib oladi)
    ValidFrom time.Time     // kurs boshlanish vaqti
    ValidTo   time.Time     // kurs amal qilish muddati
    Source    string        // kurs manbai ("CBU", "MANUAL", "API")
}
```

## Kurs Manbai

<!-- Kurslar quyidagi manbalardan olinadi -->
```
1. CBU (Markaziy bank) API — rasmiy kurs, kun boshida yangilanadi
2. Manual kirish — admin panel orqali maxsus kurs belgilash
3. Tashqi API — real-time kurslar (kelajakda)

Yangilanish tartibi:
  - CBU kurs har kuni 09:00 da avtomatik yuklanadi
  - Admin istalgan vaqtda manual kurs qo'yishi mumkin
  - Manual kurs CBU kursdan ustunlik oladi
```

## Rate Locking (Kurs qulflash)

<!-- Transfer vaqtida kurs o'zgarishi mumkin.
     Shuning uchun kurs transfer boshlanishida qulflanadi. -->
```
Transfer Flow:
  1. Foydalanuvchi transfer boshlaydi (USD → UZS)
  2. Server joriy kursni oladi va 2 daqiqaga qulflaydi
  3. Foydalanuvchiga qulflangan kurs ko'rsatiladi
  4. Foydalanuvchi tasdiqlaydi → transfer shu kurs bilan bajariladi
  5. Agar 2 daqiqa ichida tasdiqlamasa → kurs lock muddati tugaydi

Rate Lock jadvali:
  rate_lock_id    UUID
  pair            VARCHAR(7)     -- "USD/UZS"
  locked_rate     BIGINT         -- qulflangan kurs
  user_id         UUID
  expires_at      TIMESTAMPTZ    -- 2 daqiqadan keyin
  used            BOOLEAN        -- transfer da ishlatildimi
```

## FX Spread (Oldi-sotdi farqi)

```
Misol:
  CBU rasmiy kurs:  1 USD = 12,750 UZS
  Bank buy rate:    1 USD = 12,700 UZS  (bank sotib oladi)
  Bank sell rate:   1 USD = 12,800 UZS  (bank sotadi)
  Spread:           100 UZS (0.78%)

Foydalanuvchi transfer qilganda:
  - USD → UZS: sell rate ishlatiladi (12,800)
  - UZS → USD: buy rate ishlatiladi (12,700)
```

## Stale Rate Handling

<!-- Agar cache dagi kurs eskirgan bo'lsa -->
```
Qoidalar:
  - Redis cache: 5 min TTL (oddiy ko'rish uchun)
  - Transfer uchun: DOIMO DB dan yangi kurs olish (cache emas!)
  - Agar DB dagi kurs 24 soatdan eski bo'lsa:
    → Transfer BLOCK qilinadi
    → Admin alert yuboriladi
    → "Kurs yangilanmagan" xato xabari
```

## API

| Method | Endpoint | Middleware | Tavsif |
|---|---|---|---|
| GET | `/api/v1/currencies` | Public | Qo'llab-quvvatlanadigan valyutalar |
| GET | `/api/v1/currencies/rates?from=USD&to=UZS` | Public | Joriy kurs (cached) |
| POST | `/api/v1/currencies/rate-lock` | Session | Transfer uchun kurs qulflash |

Redis cache: 5 min TTL (faqat GET /rates uchun, transfer uchun emas).
