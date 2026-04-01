package transfer

import "context"

// EventRepository - event store for the Transfer aggregate
type EventRepository interface {
	// Append persists new events with optimistic concurrency.
	// If stored version != expectedVersion, returns concurrency error.
	Append(ctx context.Context, aggregateID string, expectedVersion int, events []Event) error

	// LoadEvents returns all events for an aggregate, ordered by version ASC.
	LoadEvents(ctx context.Context, aggregateID string) ([]Event, error)
}
