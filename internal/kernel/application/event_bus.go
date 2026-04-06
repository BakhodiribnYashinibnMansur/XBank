package application

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

// EventBus publishes domain events to subscribers.
// Implementations may be in-memory (sync) or distributed (Kafka, Redis Streams).
type EventBus interface {
	// Publish sends domain events to all registered subscribers.
	Publish(ctx context.Context, events ...domain.DomainEvent) error
}

// EventHandler processes a single domain event.
type EventHandler interface {
	// Handle processes the event. Return error to signal failure (for retry/DLQ).
	Handle(ctx context.Context, event domain.DomainEvent) error
}

// EventSubscriber listens for specific event types.
type EventSubscriber interface {
	// Subscribe registers a handler for the given event name.
	Subscribe(eventName string, handler EventHandler)
}
