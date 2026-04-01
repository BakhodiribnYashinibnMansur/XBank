package mock

import (
	"context"
	"sync"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/transfer"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
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

func (r *TransferEventRepository) Append(ctx context.Context, aggregateID string, expectedVersion int, events []transfer.Event) error {
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

func (r *TransferEventRepository) LoadEvents(ctx context.Context, aggregateID string) ([]transfer.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.events[aggregateID], nil
}
