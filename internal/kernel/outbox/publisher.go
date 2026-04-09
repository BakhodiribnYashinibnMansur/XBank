package outbox

import (
	"context"
)

// Publisher implements domain.EventPublisher by writing to the outbox table
// within the current database transaction instead of publishing directly to Kafka.
// This guarantees that domain events are persisted atomically with the business operation.
type Publisher struct {
	repo *Repository
}

// NewPublisher creates an outbox-backed event publisher.
func NewPublisher(repo *Repository) *Publisher {
	return &Publisher{repo: repo}
}

// Publish inserts a message into the outbox table.
// Must be called within a database transaction (ctx must carry a pgx.Tx via ExtractDBTX).
// Use WithOutboxMeta(ctx, aggregateType, aggregateID) before calling to set aggregate info.
func (p *Publisher) Publish(ctx context.Context, topic string, key string, payload []byte) error {
	meta, _ := ExtractOutboxMeta(ctx)

	return p.repo.Insert(ctx, Entry{
		AggregateType: meta.AggregateType,
		AggregateID:   meta.AggregateID,
		Topic:         topic,
		PartitionKey:  key,
		Payload:       payload,
	})
}
