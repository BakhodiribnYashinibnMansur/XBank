# Event Sourcing

## Konsept
```
Oddiy yondashuv:          Event Sourcing:
┌─────────┐               ┌─────────────────────────┐
│ balance  │               │ Event 1: +1000 (deposit) │
│  = 500   │               │ Event 2: -200 (transfer) │
└─────────┘               │ Event 3: -300 (payment)  │
                          │ Current state = 500       │
                          └─────────────────────────┘
```

- Har bir o'zgarish event sifatida saqlanadi
- Hech qachon UPDATE/DELETE — faqat APPEND
- Istalgan vaqtga "qaytib" holat ko'rish mumkin
- Audit uchun ideal — regulyator talab qiladi

## Event Turlari

<!-- Barcha Account aggregate event turlari -->

| Event | Tavsif | Ma'lumot (event_data) |
|---|---|---|
| `AccountOpenedEvent` | Yangi hisob ochildi | `{user_id, currency, account_number}` |
| `AccountCreditedEvent` | Hisobga pul tushdi | `{amount, reference, source}` |
| `AccountDebitedEvent` | Hisobdan pul yechildi | `{amount, reference, destination}` |
| `HoldPlacedEvent` | Pul bloklandi (karta auth) | `{amount, reference, expires_at}` |
| `HoldCapturedEvent` | Hold tasdiqlandi (to'lov) | `{reference, captured_amount, original_amount}` |
| `HoldReleasedEvent` | Hold bekor qilindi | `{reference, released_amount}` |
| `AccountFrozenEvent` | Hisob muzlatildi | `{reason, frozen_by}` |
| `AccountUnfrozenEvent` | Hisob qayta faollashdi | `{reason, unfrozen_by}` |
| `AccountClosedEvent` | Hisob yopildi | `{reason, closed_by}` |

### Event Strukturasi (Go)
```go
// Har bir event quyidagi interfeys'ni implement qiladi
type DomainEvent interface {
    EventType() string      // "AccountCreditedEvent"
    AggregateID() uuid.UUID // account UUID
    Version() int           // monoton o'suvchi raqam
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
    version        INTEGER NOT NULL,           -- aggregate versiyasi (1, 2, 3, ...)
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(aggregate_id, version)              -- bitta aggregate uchun version takrorlanmaydi
) PARTITION BY RANGE (created_at);

-- Indekslar
CREATE INDEX idx_event_store_aggregate ON event_store (aggregate_id, version);
CREATE INDEX idx_event_store_type ON event_store (event_type, created_at);
```

## Event Versioning (Schema Evolution)

<!-- Event strukturasi vaqt o'tishi bilan o'zgarishi mumkin.
     Masalan, AccountCreditedEvent ga yangi field qo'shilishi kerak.
     Eski eventlarni buzmasdan yangi field qo'shish strategiyasi: -->

```
Strategiya: Upcasting (o'qish vaqtida eski formatni yangi formatga o'tkazish)

v1: { "amount": 100000 }
v2: { "amount": 100000, "source": "TRANSFER" }    ← yangi field qo'shildi

Qoidalar:
1. Event data ga yangi field qo'shish — OK (default qiymat bilan)
2. Mavjud field ni o'chirish — MUMKIN EMAS
3. Field nomini o'zgartirish — MUMKIN EMAS
4. Yangi event turi yaratish — OK (eski event turi saqlanadi)

Go da:
  func (e *AccountCreditedEvent) Upcast(data map[string]any) {
      if _, ok := data["source"]; !ok {
          data["source"] = "UNKNOWN"  // eski eventlar uchun default
      }
  }
```

## Snapshot (Performance)

Har 100 eventdan keyin joriy holatni cache:
```sql
CREATE TABLE snapshots (
    aggregate_id   UUID NOT NULL,
    aggregate_type VARCHAR(50) NOT NULL,
    version        INTEGER NOT NULL,           -- shu versiongacha bo'lgan eventlar
    state          JSONB NOT NULL,             -- aggregate ning joriy holati
    checksum       VARCHAR(64),                -- SHA-256 (integrity tekshiruv)
    created_at     TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (aggregate_id, version)
);
```

### Snapshot Validatsiya
<!-- Reconciliation vaqtida snapshot to'g'riligini tekshirish -->
```
1. Snapshot yuklash (version=N, state=S1)
2. Barcha eventlarni 1 dan N gacha replay → state S2
3. S1 == S2? → OK
4. S1 != S2? → Snapshot buzilgan, qayta yaratish + ALERT
   - Buzilgan snapshot o'chiriladi
   - Barcha eventlardan qayta build qilinadi
   - Yangi snapshot saqlanadi
```

## Account Yuklash Strategiyasi
```
AccountRepository.FindByID(id):
  │
  ├── 1. Oxirgi snapshot olish (version=100)
  │       → snapshot topilmasa, version=0 dan boshlash
  │
  ├── 2. Undan keyingi eventlar (101, 102, ...)
  │       → eventlar topilmasa, snapshot holati = joriy holat
  │
  ├── 3. Snapshot + replay = joriy holat
  │
  └── 4. Agar yangi version % 100 == 0 → yangi snapshot saqlash
```

## Temporal Query (Vaqtga qarab holat ko'rish)

<!-- Regulyator yoki audit uchun: "2026-01-15 dagi balans qancha edi?" -->
```
AccountRepository.FindByIDAtTime(id, targetTime):
  1. targetTime gacha bo'lgan oxirgi snapshot olish
  2. Snapshot dan targetTime gacha bo'lgan eventlarni replay
  3. Natija = o'sha vaqtdagi aniq holat

Foydalanish holatlari:
  - Audit so'rovi: "Shu sanadagi balans?"
  - Dispute: "Transfer vaqtidagi holat?"
  - Regulyator hisoboti: "Oy oxiridagi balans?"
```

## CQRS Projections

Event store → denormalized read model (materialized views):

| Projection | Maqsad | Yangilanish |
|---|---|---|
| `AccountSummaryView` | Hisob holati: balance, status, oxirgi tx | Har bir account event da |
| `TransactionHistoryView` | Paginated tranzaksiya tarixi | Credit/Debit/Transfer event da |
| `DashboardView` | Barcha hisoblar + umumiy balance | Har bir balance o'zgarishda |

### Projection Yangilanish Mexanizmi
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
            └── Consumer: ProjectionRebuilder (agar projection buzilsa)
```

### Projection Rebuild
<!-- Agar projection noto'g'ri bo'lsa yoki yangi projection qo'shilsa -->
```
1. Yangi projection jadval yaratish
2. event_store dan BARCHA eventlarni ketma-ket o'qish
3. Har bir event ni projection handler ga berish
4. Eski projection bilan almashtirish (atomic swap)
```
