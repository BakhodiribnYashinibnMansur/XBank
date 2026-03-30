# Notifications & Audit Trail

## Real-Time Notifications (SSE)
Server-Sent Events orqali jonli bildirishnomalar:

### Event Types
- `transfer.completed` → "Sizga $250 o'tkazildi"
- `transfer.received` → "Hisobingizga $250 tushdi"
- `card.blocked` → "Kartangiz bloklandi"
- `aml.flagged` → "Tranzaksiyangiz tekshirilmoqda"
- `session.new_login` → "Yangi qurilmadan kirish aniqlandi"

### Flow
```
Domain Events → EventBus → NotificationHandler → Redis Pub/Sub → SSE Stream
```

### Notification Table
```sql
CREATE TABLE notifications (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    title      VARCHAR(200) NOT NULL,
    body       TEXT,
    is_read    BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

## Audit Trail (Immutable)
- `audit_log` jadvaliga faqat INSERT — UPDATE/DELETE yo'q
- Kim, qachon, nima qildi, qanday natija
- IP address + User-Agent + Device ID + Correlation-ID
- Retention: 7 yil (regulyator talabi)
- Partitioned by month

### Audit Log Schema
```sql
CREATE TABLE audit_log (
    id             UUID PRIMARY KEY,
    actor_id       UUID,
    actor_type     VARCHAR(20),      -- USER, SYSTEM, ADMIN
    action         VARCHAR(100),     -- TRANSFER_CREATED
    resource_type  VARCHAR(50),      -- TRANSACTION, ACCOUNT
    resource_id    UUID,
    old_value      JSONB,
    new_value      JSONB,
    ip_address     INET,
    user_agent     TEXT,
    device_id      VARCHAR(255),
    correlation_id VARCHAR(64),
    risk_score     SMALLINT,
    created_at     TIMESTAMPTZ DEFAULT NOW()
) PARTITION BY RANGE (created_at);
```

## SSE Reconnection

<!-- EventSource avtomatik qayta ulanadi, lekin ba'zi holatlarda
     qo'shimcha boshqarish kerak -->
```
Server tomonda:
  - retry: 3000 (3 sekund) field SSE event da yuboriladi
  - Last-Event-ID header orqali oxirgi event ID ni olish
  - O'tkazib yuborilgan eventlarni qayta yuborish

Client tomonda:
  - EventSource.onerror → avtomatik reconnect
  - Agar 3 marta ketma-ket fail → manual reconnect (30 sekund kutish)
```

## Notification Retention (Saqlash muddati)

<!-- Bildirishnomalar abadiy saqlanmaydi -->
```
Qoidalar:
  - Notification jadvalida: 90 kun saqlash
  - 90 kundan eski notification lar → arxivga o'tkazish (cold storage)
  - Audit log: 7 yil saqlash (regulyator talabi, alohida jadval)
  - pg_cron: har hafta eski notification larni tozalash
```

## API Endpoints

| Method | Endpoint | Middleware | Tavsif |
|---|---|---|---|
| GET | `/api/v1/notifications/stream` | Session (SSE) | Real-time bildirishnomalar stream |
| GET | `/api/v1/notifications` | Session | Bildirishnomalar ro'yxati (paginated) |
| PATCH | `/api/v1/notifications/{id}/read` | Session | Bildirishnomani o'qilgan deb belgilash |
| PATCH | `/api/v1/notifications/read-all` | Session | Barcha bildirishnomalarni o'qilgan deb belgilash |
