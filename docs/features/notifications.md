# Notifications & Audit Trail

## Real-Time Notifications (SSE)
Real-time notifications via Server-Sent Events:

### Event Types
- `transfer.completed` → "You have sent $250"
- `transfer.received` → "Your account received $250"
- `card.blocked` → "Your card has been blocked"
- `aml.flagged` → "Your transaction is being reviewed"
- `session.new_login` → "Login from a new device detected"

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
- `audit_log` table only supports INSERT — no UPDATE/DELETE
- Who, when, what action, what result
- IP address + User-Agent + Device ID + Correlation-ID
- Retention: 7 years (regulatory requirement)
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

<!-- EventSource reconnects automatically, but in some cases
     additional handling is needed -->
```
Server side:
  - retry: 3000 (3 seconds) field is sent in the SSE event
  - Retrieve the last event ID via Last-Event-ID header
  - Resend missed events

Client side:
  - EventSource.onerror → automatic reconnect
  - If 3 consecutive failures → manual reconnect (wait 30 seconds)
```

## Notification Retention

<!-- Notifications are not stored forever -->
```
Rules:
  - Notification table: 90 days retention
  - Notifications older than 90 days → moved to archive (cold storage)
  - Audit log: 7 years retention (regulatory requirement, separate table)
  - pg_cron: clean up old notifications weekly
```

## API Endpoints

| Method | Endpoint | Middleware | Description |
|---|---|---|---|
| GET | `/api/v1/notifications/stream` | Session (SSE) | Real-time notifications stream |
| GET | `/api/v1/notifications` | Session | Notifications list (paginated) |
| PATCH | `/api/v1/notifications/{id}/read` | Session | Mark notification as read |
| PATCH | `/api/v1/notifications/read-all` | Session | Mark all notifications as read |
