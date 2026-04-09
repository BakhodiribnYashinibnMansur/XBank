package mock

import (
	"context"
	"sync"

	transfer "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/domain"
)

type TransferEventRepository struct {
	mu     sync.RWMutex
	events map[string][]transfer.Event // aggregateID → events
}

func NewTransferEventRepository() *TransferEventRepository {
	return &TransferEventRepository{
		events: make(map[string][]transfer.Event),
	}
}

func (r *TransferEventRepository) Append(ctx context.Context, aggregateID string, events []transfer.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.events[aggregateID] = append(r.events[aggregateID], events...)
	return nil
}

func (r *TransferEventRepository) LoadEvents(ctx context.Context, aggregateID string) ([]transfer.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.events[aggregateID], nil
}

func (r *TransferEventRepository) LoadEventsFromVersion(ctx context.Context, aggregateID string, fromVersion int) ([]transfer.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []transfer.Event
	for _, e := range r.events[aggregateID] {
		if e.Version > fromVersion {
			result = append(result, e)
		}
	}
	return result, nil
}

func (r *TransferEventRepository) SaveSnapshot(ctx context.Context, snapshot transfer.Snapshot) error {
	return nil
}

func (r *TransferEventRepository) LoadSnapshot(ctx context.Context, aggregateID string) (*transfer.Snapshot, error) {
	return nil, nil
}
