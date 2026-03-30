# Message Queue & Async Processing

## Redis Streams (Monolith ichida)

<!-- Redis Streams — Kafka'ga o'xshash, lekin soddaroq.
     Monolith ichida async processing uchun ishlatiladi.
     Kelajakda microservice ga o'tganda Kafka/NATS ga almashtirish oson. -->

### Sinxron (real-time, HTTP javob kutadi)
- Balance check, login, 2FA verify
- Transfer saga (barcha 10 step bitta request ichida)

### Asinxron (queue orqali, background processing)
- Notification yuborish
- Fraud analysis (chuqur tekshiruv, background)
- Statement generation
- Reconciliation
- AML screening (batch)
- Projection rebuild

## Topics (Streams)

```
xbank:transfers:created   → fraud (deep scan), notification, statement
xbank:transfers:completed → analytics, reporting
xbank:transfers:failed    → alert, retry (agar retryable bo'lsa)
xbank:users:kyc:updated   → compliance, account status update
xbank:accounts:frozen     → notification, admin alert
```

## Consumer Groups

<!-- Consumer group — bir nechta consumer bitta stream'ni parallel o'qishi.
     Har bir xabar faqat BITTA consumer ga beriladi (load balancing). -->

```
Redis komandalar:
  XGROUP CREATE xbank:transfers:created fraud-group $ MKSTREAM
  XGROUP CREATE xbank:transfers:created notification-group $ MKSTREAM

  XREADGROUP GROUP fraud-group consumer-1 COUNT 10 BLOCK 5000
              STREAMS xbank:transfers:created >
```

### Consumer Group Konfiguratsiya

| Stream | Consumer Group | Consumers | Maqsad |
|---|---|---|---|
| `xbank:transfers:created` | `fraud-group` | 2 | Fraud deep analysis |
| `xbank:transfers:created` | `notification-group` | 1 | SSE notification |
| `xbank:transfers:completed` | `analytics-group` | 1 | Analytics/reporting |
| `xbank:transfers:failed` | `alert-group` | 1 | Admin alert |
| `xbank:users:kyc:updated` | `compliance-group` | 1 | KYC status sync |

## Message Format

```json
{
  "id": "msg-uuid",
  "type": "transfer.created",
  "data": {
    "transfer_id": "txn-uuid",
    "from_account": "acc-uuid",
    "to_account": "acc-uuid",
    "amount": 100000,
    "currency": "UZS"
  },
  "metadata": {
    "correlation_id": "corr-uuid",
    "user_id": "user-uuid",
    "timestamp": "2026-03-30T10:00:00Z"
  },
  "retry_count": 0
}
```

## Message Ordering Kafolatlari

<!-- Redis Streams har bir stream ichida tartibni kafolatlaydi -->
```
Kafolatlar:
  - Bitta stream ichida — FIFO tartib (yozilgan ketma-ketlikda)
  - Turli stream lar o'rtasida — tartib kafolatlanmaydi
  - Consumer group ichida — har bir xabar FAQAT bitta consumer ga

Muhim: Transfer eventlari bitta stream'da bo'lgani uchun,
       bitta account uchun eventlar tartibda keladi.
       Lekin turli account lar uchun eventlar parallel ishlaydi.
```

## Retry Mexanizmi

```
Consumer xabarni o'qidi → ishlashga harakat qildi → FAIL

Retry policy (exponential backoff):
  1-chi urinish:  1 sekund kutish   → qayta ishlash
  2-chi urinish:  5 sekund kutish   → qayta ishlash
  3-chi urinish:  30 sekund kutish  → qayta ishlash
  4-chi urinish:  → DLQ ga yuborish (manual review)

Go pseudocode:
  for attempt := 1; attempt <= 3; attempt++ {
      err := processMessage(msg)
      if err == nil {
          XACK(stream, group, msg.ID)  // muvaffaqiyat
          return
      }
      sleep(retryDelay[attempt])  // 1s, 5s, 30s
  }
  moveToDLQ(msg, lastError)  // 3 marta fail → DLQ
```

## Dead Letter Queue (DLQ)

<!-- DLQ — 3 marta retry dan keyin ham fail bo'lgan xabarlar.
     HECH QACHON avtomatik o'chirilmaydi — admin manual ko'rib chiqadi. -->

```sql
CREATE TABLE dead_letter_queue (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic       VARCHAR(100) NOT NULL,        -- qaysi stream dan kelgan
    payload     JSONB NOT NULL,               -- original xabar
    error       TEXT NOT NULL,                -- oxirgi xato xabari
    retries     INTEGER DEFAULT 0,            -- necha marta urinildi
    max_retries INTEGER DEFAULT 3,
    status      VARCHAR(20) DEFAULT 'PENDING', -- PENDING, REPROCESSED, DISCARDED
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    processed_at TIMESTAMPTZ                  -- admin qayta ishlagan vaqt
);

CREATE INDEX idx_dlq_status ON dead_letter_queue (status, created_at);
```

### DLQ Ishlash Jarayoni
```
1. Admin panel → DLQ ro'yxatini ko'rish
2. Har bir xabar uchun: payload, error, retry soni
3. Admin tanlov:
   a. "Qayta ishlash" → xabarni qayta queue ga yuborish
   b. "Bekor qilish"  → status = DISCARDED (sababini yozish)
4. Audit log ga qayd qilish
```

## Monitoring

<!-- Consumer lag va DLQ ni kuzatish juda muhim -->
```
Prometheus Metrikalar:
  - redis_stream_length          — stream dagi xabar soni (gauge)
  - redis_stream_consumer_lag    — consumer orqada qolishi (gauge)
  - redis_dlq_count              — DLQ dagi xabar soni (gauge)
  - message_processing_duration  — xabar ishlash vaqti (histogram)
  - message_processing_errors    — xato soni (counter)

Alert Qoidalari:
  - consumer_lag > 1000           → "Consumer orqada qoldi" alert
  - dlq_count > 10                → "DLQ to'lmoqda" alert
  - processing_duration_p99 > 5s  → "Sekin processing" alert
```
