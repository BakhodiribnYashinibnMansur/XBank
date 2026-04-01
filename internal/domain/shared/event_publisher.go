package shared

import "context"

// EventPublisher - interface for publishing domain events to a message broker.
// Implementation lives in infrastructure layer (e.g. Kafka).
type EventPublisher interface {
	// Publish sends a domain event to the given topic.
	// key is used for partitioning (e.g. account_id).
	// payload is a serialized message (e.g. Protobuf bytes).
	Publish(ctx context.Context, topic string, key string, payload []byte) error
}
