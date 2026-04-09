package domain

import (
	"context"
	"time"
)

// EventRepository - event store for the Account aggregate
type EventRepository interface {
	// Append persists new events. Concurrency is controlled via pessimistic locking (FOR UPDATE).
	Append(ctx context.Context, aggregateID string, events []Event) error

	// LoadEvents returns all events for an aggregate, ordered by version ASC.
	LoadEvents(ctx context.Context, aggregateID string) ([]Event, error)

	// LoadEventsFromVersion returns events after fromVersion (exclusive).
	LoadEventsFromVersion(ctx context.Context, aggregateID string, fromVersion int) ([]Event, error)

	// SaveSnapshot persists a point-in-time snapshot (upsert).
	SaveSnapshot(ctx context.Context, snapshot Snapshot) error

	// LoadSnapshot returns the latest snapshot for an aggregate (nil if none).
	LoadSnapshot(ctx context.Context, aggregateID string) (*Snapshot, error)
}

// Snapshot - materialized account state at a given version
type Snapshot struct {
	AggregateID string
	Version     int
	State       SnapshotState
	CreatedAt   time.Time
}

// SnapshotState - serializable account state for snapshots
type SnapshotState struct {
	UserID        string `json:"user_id"`
	AccountNumber string `json:"account_number"`
	Balance       int64  `json:"balance"`
	Currency      string `json:"currency"`
	Status        string `json:"status"`
}
