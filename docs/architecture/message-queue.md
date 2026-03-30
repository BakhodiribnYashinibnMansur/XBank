# Message Queue & Async Processing

## Apache Kafka

<!-- Kafka — distributed event streaming platform.
     Barcha async xabarlar Protobuf formatda serializatsiya qilinadi.
     High throughput, durability, ordering garantiyasi.
     Schema Registry orqali backward/forward compatibility. -->

### Sinxron (real-time, HTTP javob kutadi)
- Balance check, login, 2FA verify
- Transfer saga (barcha 10 step bitta request ichida)

### Asinxron (Kafka orqali, background processing)
- Notification yuborish
- Fraud analysis (chuqur tekshiruv, background)
- Statement generation
- Reconciliation
- AML screening (batch)
- Projection rebuild

## Kafka Cluster Konfiguratsiya

```yaml
# docker-compose.yml (development)
kafka:
  image: confluentinc/cp-kafka:7.7
  environment:
    KAFKA_BROKER_ID: 1
    KAFKA_NUM_PARTITIONS: 6
    KAFKA_DEFAULT_REPLICATION_FACTOR: 1          # dev=1, prod=3
    KAFKA_LOG_RETENTION_HOURS: 168               # 7 kun
    KAFKA_LOG_SEGMENT_BYTES: 1073741824          # 1 GB
    KAFKA_MESSAGE_MAX_BYTES: 1048576             # 1 MB max message
    KAFKA_AUTO_CREATE_TOPICS_ENABLE: "false"     # manual topic creation

schema-registry:
  image: confluentinc/cp-schema-registry:7.7
  environment:
    SCHEMA_REGISTRY_KAFKASTORE_BOOTSTRAP_SERVERS: kafka:9092
    SCHEMA_REGISTRY_SCHEMA_COMPATIBILITY_LEVEL: BACKWARD
```

## Topics

<!-- Topic = Kafka'dagi xabar kanali.
     Har bir topic — partition larga bo'lingan.
     Partition ichida tartib (ordering) kafolatlanadi. -->

```
xbank.transfers.created   → fraud (deep scan), notification, statement
xbank.transfers.completed → analytics, reporting
xbank.transfers.failed    → alert, retry (agar retryable bo'lsa)
xbank.users.kyc.updated   → compliance, account status update
xbank.accounts.frozen     → notification, admin alert
```

### Topic Konfiguratsiya

| Topic | Partitions | Replication | Retention | Key |
|---|---|---|---|---|
| `xbank.transfers.created` | 6 | 3 | 7 kun | `account_id` |
| `xbank.transfers.completed` | 3 | 3 | 30 kun | `transfer_id` |
| `xbank.transfers.failed` | 3 | 3 | 30 kun | `transfer_id` |
| `xbank.users.kyc.updated` | 3 | 3 | 30 kun | `user_id` |
| `xbank.accounts.frozen` | 3 | 3 | 30 kun | `account_id` |

```bash
# Topic yaratish (manual)
kafka-topics --create --topic xbank.transfers.created \
  --partitions 6 --replication-factor 3 \
  --config retention.ms=604800000 \
  --config cleanup.policy=delete \
  --bootstrap-server kafka:9092
```

## Protocol Buffers (Protobuf) — Message Format

<!-- Protobuf — Google tomonidan yaratilgan binary serialization format.
     JSON dan 3-10x kichikroq, 20-100x tezroq parse.
     Schema Registry orqali versioning va compatibility tekshiruvi. -->

### Proto Fayllar Strukturasi

```
proto/
├── common/
│   └── metadata.proto         # Umumiy metadata (correlation_id, timestamp)
├── transfers/
│   ├── transfer_created.proto
│   ├── transfer_completed.proto
│   └── transfer_failed.proto
├── users/
│   └── kyc_updated.proto
└── accounts/
    └── account_frozen.proto
```

### Umumiy Metadata

```protobuf
// proto/common/metadata.proto
syntax = "proto3";
package xbank.common;
option go_package = "github.com/BakhodiribnYashinibnMansur/XBank/pkg/proto/common";

import "google/protobuf/timestamp.proto";

message EventMetadata {
  string event_id       = 1;   // UUID — har bir xabar uchun unikal
  string correlation_id = 2;   // Request tracing uchun
  string user_id        = 3;   // Kim trigger qildi
  google.protobuf.Timestamp timestamp = 4;
  int32 retry_count     = 5;   // Necha marta qayta urinildi
  string source         = 6;   // Qaysi service dan keldi (e.g. "transfer-service")
}
```

### Transfer Created Event

```protobuf
// proto/transfers/transfer_created.proto
syntax = "proto3";
package xbank.transfers;
option go_package = "github.com/BakhodiribnYashinibnMansur/XBank/pkg/proto/transfers";

import "common/metadata.proto";

message TransferCreated {
  xbank.common.EventMetadata metadata = 1;

  string transfer_id  = 2;   // UUID
  string from_account = 3;   // Source account UUID
  string to_account   = 4;   // Target account UUID
  int64  amount       = 5;   // Minor unit (tiyin/cent) — 100000 = 1000.00 UZS
  string currency     = 6;   // ISO 4217 (e.g. "UZS", "USD")
  TransferType type   = 7;

  enum TransferType {
    TRANSFER_TYPE_UNSPECIFIED = 0;
    TRANSFER_TYPE_INTERNAL    = 1;  // Bank ichida
    TRANSFER_TYPE_EXTERNAL    = 2;  // Boshqa bank
    TRANSFER_TYPE_SCHEDULED   = 3;  // Rejalashtirilgan
  }
}
```

### Transfer Completed Event

```protobuf
// proto/transfers/transfer_completed.proto
syntax = "proto3";
package xbank.transfers;
option go_package = "github.com/BakhodiribnYashinibnMansur/XBank/pkg/proto/transfers";

import "common/metadata.proto";
import "google/protobuf/timestamp.proto";

message TransferCompleted {
  xbank.common.EventMetadata metadata = 1;

  string transfer_id  = 2;
  string from_account = 3;
  string to_account   = 4;
  int64  amount       = 5;
  string currency     = 6;
  google.protobuf.Timestamp completed_at = 7;
}
```

### Transfer Failed Event

```protobuf
// proto/transfers/transfer_failed.proto
syntax = "proto3";
package xbank.transfers;
option go_package = "github.com/BakhodiribnYashinibnMansur/XBank/pkg/proto/transfers";

import "common/metadata.proto";

message TransferFailed {
  xbank.common.EventMetadata metadata = 1;

  string transfer_id = 2;
  string from_account = 3;
  string to_account   = 4;
  int64  amount       = 5;
  string currency     = 6;
  FailureReason reason = 7;
  string error_message = 8;
  bool   retryable     = 9;

  enum FailureReason {
    FAILURE_REASON_UNSPECIFIED       = 0;
    FAILURE_REASON_INSUFFICIENT_FUNDS = 1;
    FAILURE_REASON_DAILY_LIMIT       = 2;
    FAILURE_REASON_FRAUD_DETECTED    = 3;
    FAILURE_REASON_AML_BLOCKED       = 4;
    FAILURE_REASON_ACCOUNT_FROZEN    = 5;
    FAILURE_REASON_TIMEOUT           = 6;
    FAILURE_REASON_INTERNAL_ERROR    = 7;
  }
}
```

### KYC Updated Event

```protobuf
// proto/users/kyc_updated.proto
syntax = "proto3";
package xbank.users;
option go_package = "github.com/BakhodiribnYashinibnMansur/XBank/pkg/proto/users";

import "common/metadata.proto";

message KYCUpdated {
  xbank.common.EventMetadata metadata = 1;

  string user_id    = 2;
  KYCStatus status  = 3;
  KYCLevel level    = 4;

  enum KYCStatus {
    KYC_STATUS_UNSPECIFIED = 0;
    KYC_STATUS_PENDING     = 1;
    KYC_STATUS_VERIFIED    = 2;
    KYC_STATUS_REJECTED    = 3;
  }

  enum KYCLevel {
    KYC_LEVEL_UNSPECIFIED = 0;
    KYC_LEVEL_BASIC       = 1;   // Passport only
    KYC_LEVEL_ENHANCED    = 2;   // Passport + selfie + address proof
    KYC_LEVEL_FULL        = 3;   // Video verification
  }
}
```

### Account Frozen Event

```protobuf
// proto/accounts/account_frozen.proto
syntax = "proto3";
package xbank.accounts;
option go_package = "github.com/BakhodiribnYashinibnMansur/XBank/pkg/proto/accounts";

import "common/metadata.proto";

message AccountFrozen {
  xbank.common.EventMetadata metadata = 1;

  string account_id  = 2;
  string user_id     = 3;
  FreezeReason reason = 4;
  string description  = 5;   // Admin izoh

  enum FreezeReason {
    FREEZE_REASON_UNSPECIFIED   = 0;
    FREEZE_REASON_FRAUD         = 1;
    FREEZE_REASON_AML           = 2;
    FREEZE_REASON_COURT_ORDER   = 3;
    FREEZE_REASON_USER_REQUEST  = 4;
    FREEZE_REASON_ADMIN         = 5;
  }
}
```

## Go Protobuf Generatsiya

```bash
# protoc o'rnatish va Go plugin
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Barcha proto fayllarni kompilatsiya qilish
protoc --go_out=. --go_opt=paths=source_relative \
  proto/common/*.proto \
  proto/transfers/*.proto \
  proto/users/*.proto \
  proto/accounts/*.proto
```

```makefile
# Makefile
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		proto/**/*.proto
```

## Consumer Groups

<!-- Kafka Consumer Group — bir nechta consumer bitta topic'ni parallel o'qishi.
     Har bir partition faqat BITTA consumer ga tayinlanadi (group ichida).
     Partition soni >= consumer soni bo'lishi kerak. -->

### Consumer Group Konfiguratsiya

| Topic | Consumer Group | Consumers | Maqsad |
|---|---|---|---|
| `xbank.transfers.created` | `fraud-group` | 2 | Fraud deep analysis |
| `xbank.transfers.created` | `notification-group` | 1 | SSE notification |
| `xbank.transfers.completed` | `analytics-group` | 1 | Analytics/reporting |
| `xbank.transfers.failed` | `alert-group` | 1 | Admin alert + retry |
| `xbank.users.kyc.updated` | `compliance-group` | 1 | KYC status sync |
| `xbank.accounts.frozen` | `notification-group` | 1 | User/Admin notification |

### Consumer Konfiguratsiya (Go)

```go
// Kafka consumer config
config := kafka.ReaderConfig{
    Brokers:        []string{"kafka:9092"},
    GroupID:        "fraud-group",
    Topic:          "xbank.transfers.created",
    MinBytes:       1,              // 1 byte — darhol o'qish
    MaxBytes:       10e6,           // 10 MB max batch
    CommitInterval: time.Second,    // Offset commit intervali
    StartOffset:    kafka.LastOffset,
    MaxWait:        3 * time.Second,
}
```

## Message Ordering Kafolatlari

<!-- Kafka partition ichida strict ordering kafolatlaydi.
     Key bo'yicha partitioning — bitta account uchun barcha eventlar
     bitta partition ga tushadi → tartib buzilmaydi. -->

```
Kafolatlar:
  - Partition ichida — strict FIFO tartib
  - Key = account_id → bitta account uchun eventlar doimo tartibda
  - Turli partition lar o'rtasida — tartib kafolatlanmaydi
  - Consumer group ichida — har bir partition FAQAT bitta consumer ga

Muhim: Transfer eventlari key = account_id bilan yuboriladi,
       shuning uchun bitta account uchun eventlar DOIMO tartibda.
       Turli account lar uchun eventlar parallel (turli partition larda).
```

### Partitioning Strategiya

```go
// Producer — key bo'yicha partition tanlash
writer := kafka.Writer{
    Addr:     kafka.TCP("kafka:9092"),
    Topic:    "xbank.transfers.created",
    Balancer: &kafka.Murmur2Balancer{}, // Kafka default partitioner
}

// Xabar yuborish
msg := kafka.Message{
    Key:   []byte(accountID),  // account_id → partition key
    Value: protoBytes,         // Protobuf serialized
    Headers: []kafka.Header{
        {Key: "event_type", Value: []byte("transfer.created")},
        {Key: "correlation_id", Value: []byte(correlationID)},
    },
}
err := writer.WriteMessages(ctx, msg)
```

## Retry Mexanizmi

```
Consumer xabarni o'qidi → deserialize (Protobuf) → ishlash → FAIL

Retry policy (exponential backoff):
  1-chi urinish:  1 sekund kutish   → qayta ishlash
  2-chi urinish:  5 sekund kutish   → qayta ishlash
  3-chi urinish:  30 sekund kutish  → qayta ishlash
  4-chi urinish:  → DLQ topic ga yuborish (manual review)

Go pseudocode:
  for attempt := 1; attempt <= 3; attempt++ {
      event := &transfers.TransferCreated{}
      if err := proto.Unmarshal(msg.Value, event); err != nil {
          moveToDLQ(msg, err)  // deserialize fail → darhol DLQ
          return
      }
      err := processEvent(ctx, event)
      if err == nil {
          reader.CommitMessages(ctx, msg)  // muvaffaqiyat → offset commit
          return
      }
      event.Metadata.RetryCount++
      sleep(retryDelay[attempt])  // 1s, 5s, 30s
  }
  moveToDLQ(msg, lastError)  // 3 marta fail → DLQ
```

### Retry Topic Pattern

```
Asosiy flow:
  xbank.transfers.created
      │
      ├── Success → commit offset
      └── Fail → retry 3x
              │
              ├── Success → commit offset
              └── 3x fail → xbank.dlq (Dead Letter Topic)
```

## Dead Letter Queue (DLQ)

<!-- DLQ — Kafka'da alohida topic + PostgreSQL jadval.
     3 marta retry dan keyin fail bo'lgan xabarlar shu yerga tushadi.
     HECH QACHON avtomatik o'chirilmaydi — admin manual ko'rib chiqadi. -->

### DLQ Kafka Topic

```
Topic:     xbank.dlq
Partitions: 1
Retention: unlimited (cleanup.policy=compact)
```

### DLQ PostgreSQL Jadvali (Audit + Admin Panel uchun)

```sql
CREATE TABLE dead_letter_queue (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic         VARCHAR(100) NOT NULL,         -- qaysi topic dan kelgan
    partition_id  INTEGER NOT NULL,              -- Kafka partition
    offset_id     BIGINT NOT NULL,               -- Kafka offset
    key           BYTEA,                          -- Kafka message key
    payload       BYTEA NOT NULL,                -- Protobuf binary (original xabar)
    payload_json  JSONB,                          -- Debug uchun JSON representation
    error         TEXT NOT NULL,                 -- oxirgi xato xabari
    retries       INTEGER DEFAULT 0,             -- necha marta urinildi
    max_retries   INTEGER DEFAULT 3,
    status        VARCHAR(20) DEFAULT 'PENDING', -- PENDING, REPROCESSED, DISCARDED
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    processed_at  TIMESTAMPTZ                    -- admin qayta ishlagan vaqt
);

CREATE INDEX idx_dlq_status ON dead_letter_queue (status, created_at);
CREATE INDEX idx_dlq_topic ON dead_letter_queue (topic, created_at);
```

### DLQ Ishlash Jarayoni

```
1. Admin panel → DLQ ro'yxatini ko'rish
2. Har bir xabar uchun: topic, payload (JSON view), error, retry soni
3. Admin tanlov:
   a. "Qayta ishlash" → xabarni qayta Kafka topic ga yuborish
   b. "Bekor qilish"  → status = DISCARDED (sababini yozish)
4. Audit log ga qayd qilish
```

## Producer/Consumer Arxitekturasi

```
Application Layer (after commit)
    │
    ├── Sync: EventBus.Publish()
    │       └── NotificationHandler → SSE push (real-time)
    │
    └── Async: Kafka Producer
            │
            ├── proto.Marshal(TransferCreated{...})
            │
            └── kafka.Writer.WriteMessages()
                    │
                    ├── xbank.transfers.created (key=account_id)
                    │       ├── Consumer Group: fraud-group
                    │       │       └── proto.Unmarshal → FraudAnalysis
                    │       └── Consumer Group: notification-group
                    │               └── proto.Unmarshal → StatementGenerator
                    │
                    └── xbank.transfers.failed (key=transfer_id)
                            └── Consumer Group: alert-group
                                    └── proto.Unmarshal → AlertService
                                            │
                                            └── Retry 3x → fail → xbank.dlq
                                                                    │
                                                                    └── Admin manual review
```

## Monitoring

<!-- Consumer lag va DLQ ni kuzatish juda muhim.
     Kafka JMX metrikalarini Prometheus orqali yig'ish. -->

```
Prometheus Metrikalar:
  # Kafka Broker
  - kafka_server_broker_topic_metrics_messages_in_total    — topic ga kelgan xabar soni
  - kafka_server_broker_topic_metrics_bytes_in_total       — topic ga kelgan byte soni

  # Consumer Lag (eng muhim!)
  - kafka_consumer_group_lag                              — consumer qancha orqada (gauge)
  - kafka_consumer_group_current_offset                   — hozirgi o'qilgan offset

  # Application Level
  - xbank_message_processing_duration_seconds             — xabar ishlash vaqti (histogram)
  - xbank_message_processing_errors_total                 — xato soni (counter)
  - xbank_message_produced_total                          — yuborilgan xabar soni (counter)
  - xbank_message_consumed_total                          — o'qilgan xabar soni (counter)
  - xbank_dlq_count                                       — DLQ dagi xabar soni (gauge)
  - xbank_protobuf_unmarshal_errors_total                 — deserialize xatolari (counter)

Alert Qoidalari:
  - consumer_group_lag > 1000          → "Consumer orqada qoldi" alert
  - dlq_count > 10                     → "DLQ to'lmoqda" alert
  - processing_duration_p99 > 5s       → "Sekin processing" alert
  - protobuf_unmarshal_errors > 0      → "Schema mismatch" alert
  - kafka_under_replicated_partitions  → "Replication muammo" alert
```

## Schema Evolution (Protobuf Versioning)

<!-- Protobuf backward compatibility qoidalari:
     - Yangi field qo'shish — OK (eski consumer ignore qiladi)
     - Field o'chirish — XAVFLI (eski consumer crash bo'lishi mumkin)
     - Field type o'zgartirish — MAN ETILADI -->

```
Qoidalar:
  1. Yangi field qo'shish                    → ✅ Xavfsiz (backward compatible)
  2. Field raqamini qayta ishlatmaslik        → ✅ Majburiy (reserved keyword)
  3. Field type o'zgartirmaslik               → ❌ Breaking change
  4. Enum ga yangi qiymat qo'shish           → ✅ Xavfsiz (UNSPECIFIED = 0 bo'lsa)
  5. Field o'chirish o'rniga reserved qilish  → ✅ To'g'ri usul

Misol — field olib tashlash:
  message TransferCreated {
    reserved 8;              // eski field raqami — qayta ishlatmaslik
    reserved "old_field";    // eski field nomi
    ...
  }
```

## Kafka vs Redis Streams

| Xususiyat | Kafka | Redis Streams |
|---|---|---|
| **Durability** | Disk-based, replication | In-memory + AOF |
| **Throughput** | Millionlab msg/sec | O'rtacha |
| **Ordering** | Partition ichida strict | Stream ichida strict |
| **Retention** | Konfiguratsiya mumkin (kunlar/hafta) | Memory limit |
| **Consumer Groups** | ✅ Native | ✅ Native |
| **Schema Registry** | ✅ Protobuf/Avro/JSON | ❌ Yo'q |
| **Replay** | ✅ Offset dan qayta o'qish | ⚠️ Cheklangan |
| **Scalability** | Horizontal (broker qo'shish) | Vertical |
| **Monitoring** | JMX + Kafka UI + Grafana | Redis CLI |

**Tanlash sababi:** Banking tizimda durability, schema evolution, va message replay muhim — Kafka mos keladi.

## Go Dependencies

```bash
# Kafka client
go get github.com/segmentio/kafka-go

# Protobuf
go get google.golang.org/protobuf

# Schema registry (optional, production uchun)
go get github.com/riferrei/srclient
```
