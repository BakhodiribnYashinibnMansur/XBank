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

<!-- Transfer saga ning barcha holatlari va o'tishlari -->
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
            │ │ │  │ COMPLETED │ ← yakuniy holat
            │ │ │  └───────────┘
            │ │ │
            ▼ ▼ ▼       (har qanday step fail bo'lsa)
       ┌────────────┐
       │COMPENSATING│ → teskari tartibda compensate
       └─────┬──────┘
             │
             ▼
       ┌──────────┐
       │  FAILED  │ ← yakuniy holat
       └──────────┘
```

## Saga State Persistence

<!-- Saga holati DB da saqlanadi — server crash bo'lsa ham davom ettirish mumkin -->
```sql
CREATE TABLE saga_state (
    id              UUID PRIMARY KEY,
    transfer_id     UUID NOT NULL UNIQUE,
    current_step    INTEGER NOT NULL DEFAULT 0,    -- hozirgi step raqami
    status          VARCHAR(20) NOT NULL,          -- CREATED, CHECKING, LOCKING, EXECUTING, COMPLETING, COMPENSATING, COMPLETED, FAILED
    step_results    JSONB DEFAULT '{}',            -- har bir step natijasi
    error           TEXT,                          -- xato xabari (agar bor bo'lsa)
    started_at      TIMESTAMPTZ DEFAULT NOW(),
    updated_at      TIMESTAMPTZ DEFAULT NOW(),
    completed_at    TIMESTAMPTZ
);

-- Stuck saga larni topish uchun
CREATE INDEX idx_saga_active ON saga_state (status, updated_at)
    WHERE status NOT IN ('COMPLETED', 'FAILED');
```

### Saga Recovery (Server Crash)
```
Server qayta ishga tushganda:
  1. saga_state dan status != COMPLETED/FAILED larni yuklash
  2. Har bir saga uchun:
     a. current_step dan davom ettirish (idempotent bo'lgani uchun xavfsiz)
     b. Agar timeout o'tgan bo'lsa → compensate
```

## Compensatsiya

Agar Step 7 fail bo'lsa (teskari tartibda):
1. Step 6 compensate: Refund Source — debit qilingan summani qaytarish
2. Step 5 compensate: Unlock both — ikkala account lock ni ochish
3. Step 4 compensate: (amal yo'q — faqat tekshiruv edi)
4. Step 3 compensate: Unlock — source account lock ni ochish
5. Transfer status = FAILED

<!-- Har bir compensatsiya ham IDEMPOTENT bo'lishi kerak —
     bir necha marta chaqirilsa ham bir xil natija berishi kerak -->

## Timeout va Retry

```
Saga Timeout:
  - Umumiy timeout: 30 sekund (barcha steplar uchun)
  - Per-step timeout: 5 sekund
  - Timeout bo'lsa → compensate boshlanadi

Retry Strategiyasi:
  - Step fail bo'lsa → 3 marta retry (100ms → 200ms → 400ms)
  - Barcha retry lar fail → compensate
  - Retryable xatolar: network error, DB connection, serialization failure
  - Not retryable: insufficient balance, account frozen, AML block
```

## Deadlock Prevention

Account'larni UUID tartibida lock qilish:
```go
// Doimo kichik UUID birinchi lock qilinadi
// Bu barcha concurrent transfer lar bir xil tartibda lock qilishini ta'minlaydi
if fromID > toID {
    lock(toID)    // kichik UUID birinchi
    lock(fromID)  // katta UUID keyin
} else {
    lock(fromID)
    lock(toID)
}
```

## Monitoring va Alerting

<!-- Saga larni monitoring qilish uchun quyidagi metrikalar kuzatiladi -->
```
Prometheus Metrikalar:
  - saga_duration_seconds        — saga bajarilish vaqti (histogram)
  - saga_step_duration_seconds   — har bir step vaqti (histogram)
  - saga_total                   — jami saga lar (counter, label: status)
  - saga_compensation_total      — compensatsiya soni (counter)
  - saga_active_count            — hozir faol saga lar soni (gauge)

Alert Qoidalari:
  - saga_active_count > 100          → "Ko'p concurrent saga" alert
  - saga_duration_seconds > 30       → "Stuck saga" alert
  - saga_compensation_total tez o'sish → "Ko'p transfer fail" alert
```
