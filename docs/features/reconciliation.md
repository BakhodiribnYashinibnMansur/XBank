# Reconciliation — Kunlik Tekshiruv

<!-- Reconciliation — moliyaviy ma'lumotlar to'g'riligini tekshirish jarayoni.
     Bank tizimida eng muhim jarayonlardan biri.
     Har kuni avtomatik va real-time tekshiruvlar o'tkaziladi. -->

## Kunlik Tekshiruvlar (har kuni 3:00 AM, pg_cron)

### 1. Double-Entry Integrity
<!-- Barcha ledger entry larning yig'indisi DOIMO 0 ga teng bo'lishi kerak.
     Agar 0 emas → pul "yo'qolgan" yoki "ortiqcha yaratilgan". -->
```sql
SELECT CASE WHEN SUM(
    CASE WHEN side='DEBIT' THEN -amount_minor ELSE amount_minor END
) = 0 THEN 'OK' ELSE 'MISMATCH!' END
FROM ledger_entries
WHERE created_at >= CURRENT_DATE - INTERVAL '1 day';
-- Kechagi kun uchun tekshirish
-- 0 bo'lishi SHART — aks holda jiddiy muammo!
```

### 2. Account Balance = Ledger Sum
<!-- Har bir account balans'i uning barcha ledger entry lari yig'indisiga teng bo'lishi kerak -->
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
-- Bo'sh natija = OK (barcha balanslar to'g'ri)
-- Qator bor = ALERT! (balance va ledger mos emas)
```

### 3. Event Store Integrity
<!-- Event sourcing da account holati = barcha eventlarni replay qilgan natija -->
```
Har bir account uchun:
  1. Account.balance (DB dagi qiymat) olish
  2. Barcha eventlarni boshidan replay qilish → hisoblangan balance
  3. DB balance == hisoblangan balance? → OK
  4. Teng emas? → ALERT! + account freeze
```

### 4. Snapshot Validation
<!-- Snapshot to'g'riligini tekshirish — snapshot ning state qiymati
     eventlarni shu versiongacha replay qilgan natijaga teng bo'lishi kerak -->
```
Har bir snapshot uchun:
  1. Snapshot yuklash (version=N, state=S1)
  2. Eventlarni 1 dan N gacha replay → state S2
  3. S1 == S2? → OK
  4. S1 != S2? → Snapshot buzilgan → qayta yaratish + ALERT
```

### 5. Transfer Status Consistency
<!-- Barcha COMPLETED transfer larda ledger entry bo'lishi kerak -->
```sql
SELECT t.id FROM transfers t
WHERE t.status = 'COMPLETED'
AND NOT EXISTS (
    SELECT 1 FROM ledger_entries le WHERE le.transfer_id = t.id
);
-- COMPLETED transfer, lekin ledger entry yo'q = JIDDIY MUAMMO
```

## Real-time Tekshiruvlar

<!-- Kunlik tekshiruvga qo'shimcha, ba'zi tekshiruvlar real-time o'tkaziladi -->
```
Har bir transfer dan keyin (inline):
  1. Source account: balance >= 0 (CHECK constraint orqali)
  2. Ledger: debit + credit = 0 (application level)
  3. Event version: monoton o'sish (UNIQUE constraint)

Har 1 soatda (pg_cron):
  1. PENDING statusdagi transfer lar > 5 daqiqa → ALERT (stuck transfer)
  2. Hold amount > 24 soat → ALERT (expired hold)
```

## Nomuvofiqlik Topilganda (Mismatch Handling)

```
Jiddiylik darajalari:

CRITICAL (avtomatik harakat):
  - Double-entry integrity fail → barcha transfer lar TO'XTATILADI
  - Audit log + admin alert + SMS
  - Manual tekshiruvgacha tizim read-only rejimga o'tadi

HIGH (account level):
  - Account balance != ledger sum → account FREEZE
  - Account balance != event replay → account FREEZE
  - Admin manual tekshiruvga yuboradi

MEDIUM (data quality):
  - Snapshot invalid → snapshot qayta yaratish (avtomatik)
  - Transfer status inconsistency → admin review queue ga

Hal qilish jarayoni:
  1. ALERT notification (admin + oncall)
  2. Audit log entry (RECONCILIATION_MISMATCH, tafsilotlar bilan)
  3. Tegishli account(lar) FREEZE (agar jiddiy)
  4. Admin tekshiruvi:
     a. Root cause aniqlash (event log, audit trail ko'rish)
     b. Tuzatish (manual adjustment yoki bug fix)
     c. Account UNFREEZE
  5. Post-mortem yozish
```

## Reconciliation Natijalari Logi

```sql
CREATE TABLE reconciliation_runs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_type        VARCHAR(20) NOT NULL,     -- DAILY, HOURLY, MANUAL
    started_at      TIMESTAMPTZ NOT NULL,
    completed_at    TIMESTAMPTZ,
    status          VARCHAR(20) NOT NULL,     -- RUNNING, PASSED, FAILED
    checks_passed   INTEGER DEFAULT 0,        -- nechta tekshiruv o'tdi
    checks_failed   INTEGER DEFAULT 0,        -- nechta tekshiruv fail
    details         JSONB,                    -- har bir tekshiruv natijasi
    accounts_frozen INTEGER DEFAULT 0,        -- nechta account muzlatildi
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_recon_runs ON reconciliation_runs (run_type, started_at DESC);
```

## Performance

<!-- Katta tizimda reconciliation sekin bo'lishi mumkin -->
```
Optimizatsiya:
  - Kunlik tekshiruv faqat kechagi kun uchun (partitioned jadvallar tufayli tez)
  - Event replay — snapshot dan boshlash (100 event birgalikda emas, 1-2 event)
  - Parallel tekshirish — har bir account alohida goroutine da
  - Haftalik to'liq tekshiruv — barcha tarix uchun (dam olish kunlari, past yuklanish)

Taxminiy vaqt:
  - 10,000 account: ~30 sekund
  - 100,000 account: ~5 daqiqa
  - 1,000,000 account: ~1 soat (parallel, snapshot bilan)
```
