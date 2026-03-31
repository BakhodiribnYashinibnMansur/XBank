# Message Queue & Async Processing

## Apache Kafka

<!-- Kafka — distributed event streaming platform.
     All async messages are serialized in Protobuf format.
     High throughput, durability, ordering guarantee.
     Backward/forward compatibility via Schema Registry. -->

### Synchronous (real-time, waits for HTTP response)
- Balance check, login, 2FA verify
- Transfer saga (all 10 steps within a single request)

### Asynchronous (via Kafka, background processing)
- Sending notifications
- Fraud analysis (deep check, background)
- Statement generation
- Reconciliation
- AML screening (batch)
- Projection rebuild

## Kafka Cluster Configuration

```yaml
# docker-compose.yml (development)
kafka:
  image: confluentinc/cp-kafka:7.7
  environment:
    KAFKA_BROKER_ID: 1
    KAFKA_NUM_PARTITIONS: 6
    KAFKA_DEFAULT_REPLICATION_FACTOR: 1          # dev=1, prod=3
    KAFKA_LOG_RETENTION_HOURS: 168               # 7 days
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

<!-- Topic = message channel in Kafka.
     Each topic is divided into partitions.
     Ordering is guaranteed within a partition. -->

```
xbank.transfers.created   → fraud (deep scan), notification, statement
xbank.transfers.completed → analytics, reporting
xbank.transfers.failed    → alert, retry (if retryable)
xbank.users.kyc.updated   → compliance, account status update
xbank.accounts.frozen     → notification, admin alert
```

### Topic Configuration

| Topic | Partitions | Replication | Retention | Key |
|---|---|---|---|---|
| `xbank.transfers.created` | 6 | 3 | 7 days | `account_id` |
| `xbank.transfers.completed` | 3 | 3 | 30 days | `transfer_id` |
| `xbank.transfers.failed` | 3 | 3 | 30 days | `transfer_id` |
| `xbank.users.kyc.updated` | 3 | 3 | 30 days | `user_id` |
| `xbank.accounts.frozen` | 3 | 3 | 30 days | `account_id` |

```bash
# Create topic (manual)
kafka-topics --create --topic xbank.transfers.created \
  --partitions 6 --replication-factor 3 \
  --config retention.ms=604800000 \
  --config cleanup.policy=delete \
  --bootstrap-server kafka:9092
```

## Protocol Buffers (Protobuf) — Message Format

<!-- Protobuf — binary serialization format created by Google.
     3-10x smaller than JSON, 20-100x faster to parse.
     Versioning and compatibility checking via Schema Registry. -->

### Proto File Structure

```
proto/
├── common/
│   └── metadata.proto         # Common metadata (correlation_id, timestamp)
├── transfers/
│   ├── transfer_created.proto
│   ├── transfer_completed.proto
│   └── transfer_failed.proto
├── users/
│   └── kyc_updated.proto
└── accounts/
    └── account_frozen.proto
```

### Common Metadata

```protobuf
// proto/common/metadata.proto
syntax = "proto3";
package xbank.common;
option go_package = "github.com/BakhodiribnYashinibnMansur/XBank/pkg/proto/common";

import "google/protobuf/timestamp.proto";

message EventMetadata {
  string event_id       = 1;   // UUID — unique for each message
  string correlation_id = 2;   // For request tracing
  string user_id        = 3;   // Who triggered it
  google.protobuf.Timestamp timestamp = 4;
  int32 retry_count     = 5;   // How many times retried
  string source         = 6;   // Which service it came from (e.g. "transfer-service")
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
    TRANSFER_TYPE_INTERNAL    = 1;  // Within the bank
    TRANSFER_TYPE_EXTERNAL    = 2;  // Another bank
    TRANSFER_TYPE_SCHEDULED   = 3;  // Scheduled
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
  string description  = 5;   // Admin note

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

## Go Protobuf Generation

```bash
# Install protoc and Go plugin
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Compile all proto files
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

<!-- Kafka Consumer Group — multiple consumers reading a single topic in parallel.
     Each partition is assigned to ONLY ONE consumer (within the group).
     Number of partitions must be >= number of consumers. -->

### Consumer Group Configuration

| Topic | Consumer Group | Consumers | Purpose |
|---|---|---|---|
| `xbank.transfers.created` | `fraud-group` | 2 | Fraud deep analysis |
| `xbank.transfers.created` | `notification-group` | 1 | SSE notification |
| `xbank.transfers.completed` | `analytics-group` | 1 | Analytics/reporting |
| `xbank.transfers.failed` | `alert-group` | 1 | Admin alert + retry |
| `xbank.users.kyc.updated` | `compliance-group` | 1 | KYC status sync |
| `xbank.accounts.frozen` | `notification-group` | 1 | User/Admin notification |

### Consumer Configuration (Go)

```go
// Kafka consumer config
config := kafka.ReaderConfig{
    Brokers:        []string{"kafka:9092"},
    GroupID:        "fraud-group",
    Topic:          "xbank.transfers.created",
    MinBytes:       1,              // 1 byte — read immediately
    MaxBytes:       10e6,           // 10 MB max batch
    CommitInterval: time.Second,    // Offset commit interval
    StartOffset:    kafka.LastOffset,
    MaxWait:        3 * time.Second,
}
```

## Message Ordering Guarantees

<!-- Kafka guarantees strict ordering within a partition.
     Partitioning by key — all events for a single account
     go to the same partition → order is preserved. -->

```
Guarantees:
  - Within a partition — strict FIFO order
  - Key = account_id → events for a single account are always in order
  - Between different partitions — order is NOT guaranteed
  - Within a consumer group — each partition goes to ONLY one consumer

Important: Transfer events are sent with key = account_id,
           so events for a single account are ALWAYS in order.
           Events for different accounts are parallel (in different partitions).
```

### Partitioning Strategy

```go
// Producer — select partition by key
writer := kafka.Writer{
    Addr:     kafka.TCP("kafka:9092"),
    Topic:    "xbank.transfers.created",
    Balancer: &kafka.Murmur2Balancer{}, // Kafka default partitioner
}

// Send message
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

## Retry Mechanism

```
Consumer reads message → deserialize (Protobuf) → process → FAIL

Retry policy (exponential backoff):
  1st attempt:  wait 1 second   → reprocess
  2nd attempt:  wait 5 seconds  → reprocess
  3rd attempt:  wait 30 seconds → reprocess
  4th attempt:  → send to DLQ topic (manual review)

Go pseudocode:
  for attempt := 1; attempt <= 3; attempt++ {
      event := &transfers.TransferCreated{}
      if err := proto.Unmarshal(msg.Value, event); err != nil {
          moveToDLQ(msg, err)  // deserialize fail → immediately to DLQ
          return
      }
      err := processEvent(ctx, event)
      if err == nil {
          reader.CommitMessages(ctx, msg)  // success → offset commit
          return
      }
      event.Metadata.RetryCount++
      sleep(retryDelay[attempt])  // 1s, 5s, 30s
  }
  moveToDLQ(msg, lastError)  // 3 failures → DLQ
```

### Retry Topic Pattern

```
Main flow:
  xbank.transfers.created
      │
      ├── Success → commit offset
      └── Fail → retry 3x
              │
              ├── Success → commit offset
              └── 3x fail → xbank.dlq (Dead Letter Topic)
```

## Dead Letter Queue (DLQ)

<!-- DLQ — a separate topic in Kafka + PostgreSQL table.
     Messages that fail after 3 retries end up here.
     NEVER automatically deleted — admin reviews manually. -->

### DLQ Kafka Topic

```
Topic:     xbank.dlq
Partitions: 1
Retention: unlimited (cleanup.policy=compact)
```

### DLQ PostgreSQL Table (For Audit + Admin Panel)

```sql
CREATE TABLE dead_letter_queue (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic         VARCHAR(100) NOT NULL,         -- which topic it came from
    partition_id  INTEGER NOT NULL,              -- Kafka partition
    offset_id     BIGINT NOT NULL,               -- Kafka offset
    key           BYTEA,                          -- Kafka message key
    payload       BYTEA NOT NULL,                -- Protobuf binary (original message)
    payload_json  JSONB,                          -- JSON representation for debugging
    error         TEXT NOT NULL,                 -- last error message
    retries       INTEGER DEFAULT 0,             -- how many times attempted
    max_retries   INTEGER DEFAULT 3,
    status        VARCHAR(20) DEFAULT 'PENDING', -- PENDING, REPROCESSED, DISCARDED
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    processed_at  TIMESTAMPTZ                    -- time when admin reprocessed
);

CREATE INDEX idx_dlq_status ON dead_letter_queue (status, created_at);
CREATE INDEX idx_dlq_topic ON dead_letter_queue (topic, created_at);
```

### DLQ Processing Workflow

```
1. Admin panel → view DLQ list
2. For each message: topic, payload (JSON view), error, retry count
3. Admin choice:
   a. "Reprocess" → send message back to the Kafka topic
   b. "Discard"   → status = DISCARDED (record the reason)
4. Record in audit log
```

## Producer/Consumer Architecture

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

<!-- Monitoring consumer lag and DLQ is very important.
     Collecting Kafka JMX metrics via Prometheus. -->

```
Prometheus Metrics:
  # Kafka Broker
  - kafka_server_broker_topic_metrics_messages_in_total    — number of messages received by topic
  - kafka_server_broker_topic_metrics_bytes_in_total       — number of bytes received by topic

  # Consumer Lag (most important!)
  - kafka_consumer_group_lag                              — how far behind the consumer is (gauge)
  - kafka_consumer_group_current_offset                   — current read offset

  # Application Level
  - xbank_message_processing_duration_seconds             — message processing time (histogram)
  - xbank_message_processing_errors_total                 — error count (counter)
  - xbank_message_produced_total                          — number of messages sent (counter)
  - xbank_message_consumed_total                          — number of messages read (counter)
  - xbank_dlq_count                                       — number of messages in DLQ (gauge)
  - xbank_protobuf_unmarshal_errors_total                 — deserialization errors (counter)

Alert Rules:
  - consumer_group_lag > 1000          → "Consumer falling behind" alert
  - dlq_count > 10                     → "DLQ filling up" alert
  - processing_duration_p99 > 5s       → "Slow processing" alert
  - protobuf_unmarshal_errors > 0      → "Schema mismatch" alert
  - kafka_under_replicated_partitions  → "Replication issue" alert
```

## Schema Evolution (Protobuf Versioning)

<!-- Protobuf backward compatibility rules:
     - Adding a new field — OK (old consumer ignores it)
     - Removing a field — DANGEROUS (old consumer may crash)
     - Changing field type — PROHIBITED -->

```
Rules:
  1. Adding a new field                     → OK (backward compatible)
  2. Not reusing field numbers              → Mandatory (reserved keyword)
  3. Not changing field type                → Breaking change
  4. Adding a new value to enum             → OK (if UNSPECIFIED = 0)
  5. Using reserved instead of deleting     → Correct approach

Example — removing a field:
  message TransferCreated {
    reserved 8;              // old field number — do not reuse
    reserved "old_field";    // old field name
    ...
  }
```

## Kafka vs Redis Streams

| Feature | Kafka | Redis Streams |
|---|---|---|
| **Durability** | Disk-based, replication | In-memory + AOF |
| **Throughput** | Millions of msg/sec | Medium |
| **Ordering** | Strict within partition | Strict within stream |
| **Retention** | Configurable (days/weeks) | Memory limit |
| **Consumer Groups** | Native | Native |
| **Schema Registry** | Protobuf/Avro/JSON | None |
| **Replay** | Re-read from offset | Limited |
| **Scalability** | Horizontal (add brokers) | Vertical |
| **Monitoring** | JMX + Kafka UI + Grafana | Redis CLI |

**Reason for choosing:** In a banking system, durability, schema evolution, and message replay are important — Kafka is a good fit.

## Go Dependencies

```bash
# Kafka client
go get github.com/segmentio/kafka-go

# Protobuf
go get google.golang.org/protobuf

# Schema registry (optional, for production)
go get github.com/riferrei/srclient
```
