// Package outbox implements the transactional outbox pattern.
//
// Instead of publishing domain events directly to Kafka after a DB commit (best-effort),
// events are inserted into the outbox table within the same database transaction.
// A background Relay worker then reads from the outbox and publishes to Kafka,
// deleting the entry after successful delivery.
//
// Components:
//   - Publisher: domain.EventPublisher that writes to the outbox table (used by services)
//   - Repository: persistence layer for the outbox table
//   - Relay: background worker that reads outbox → publishes to Kafka → deletes entry
//   - WithOutboxMeta: context helper to carry aggregate info for outbox entries
package outbox
