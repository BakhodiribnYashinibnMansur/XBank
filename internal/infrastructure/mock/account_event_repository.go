package mock

import (
	"context"
	"sync"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/account"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
)

type AccountEventRepository struct {
	mu        sync.RWMutex
	events    map[string][]account.Event    // aggregateID → events
	snapshots map[string]*account.Snapshot  // aggregateID → latest snapshot
}

func NewAccountEventRepository() *AccountEventRepository {
	return &AccountEventRepository{
		events:    make(map[string][]account.Event),
		snapshots: make(map[string]*account.Snapshot),
	}
}

func (r *AccountEventRepository) Append(ctx context.Context, aggregateID string, expectedVersion int, events []account.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing := r.events[aggregateID]
	currentVersion := 0
	if len(existing) > 0 {
		currentVersion = existing[len(existing)-1].Version
	}

	if currentVersion != expectedVersion {
		return apperror.ErrConcurrencyConflict
	}

	r.events[aggregateID] = append(existing, events...)
	return nil
}

func (r *AccountEventRepository) LoadEvents(ctx context.Context, aggregateID string) ([]account.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.events[aggregateID], nil
}

func (r *AccountEventRepository) LoadEventsFromVersion(ctx context.Context, aggregateID string, fromVersion int) ([]account.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	all := r.events[aggregateID]
	var result []account.Event
	for _, e := range all {
		if e.Version > fromVersion {
			result = append(result, e)
		}
	}
	return result, nil
}

func (r *AccountEventRepository) SaveSnapshot(ctx context.Context, snapshot account.Snapshot) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.snapshots[snapshot.AggregateID] = &snapshot
	return nil
}

func (r *AccountEventRepository) LoadSnapshot(ctx context.Context, aggregateID string) (*account.Snapshot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.snapshots[aggregateID], nil
}
