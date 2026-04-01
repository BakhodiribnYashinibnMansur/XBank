package transfer

import (
	"context"
	"time"
)

// EventRepository - event store for the Transfer aggregate
type EventRepository interface {
	// Append persists new events with optimistic concurrency.
	// If stored version != expectedVersion, returns concurrency error.
	Append(ctx context.Context, aggregateID string, expectedVersion int, events []Event) error

	// LoadEvents returns all events for an aggregate, ordered by version ASC.
	LoadEvents(ctx context.Context, aggregateID string) ([]Event, error)

	// LoadEventsFromVersion returns events after fromVersion (exclusive).
	LoadEventsFromVersion(ctx context.Context, aggregateID string, fromVersion int) ([]Event, error)

	// SaveSnapshot persists a point-in-time snapshot (upsert).
	SaveSnapshot(ctx context.Context, snapshot Snapshot) error

	// LoadSnapshot returns the latest snapshot for an aggregate (nil if none).
	LoadSnapshot(ctx context.Context, aggregateID string) (*Snapshot, error)
}

// Snapshot - materialized transfer state at a given version
type Snapshot struct {
	AggregateID string
	Version     int
	State       SnapshotState
	CreatedAt   time.Time
}

// SnapshotState - serializable transfer state for snapshots
type SnapshotState struct {
	FromAccountID string `json:"from_account_id"`
	ToAccountID   string `json:"to_account_id"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	Status        string `json:"status"`
	Description   string `json:"description"`
	FailureReason string `json:"failure_reason"`
}
