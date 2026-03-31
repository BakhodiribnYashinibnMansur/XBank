package mock

import (
	"context"
	"sync"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/transfer"
	"github.com/google/uuid"
)

type TransferRepository struct {
	mu        sync.RWMutex
	transfers map[string]*transfer.Transfer
}

func NewTransferRepository() *TransferRepository {
	return &TransferRepository{transfers: make(map[string]*transfer.Transfer)}
}

func (r *TransferRepository) Create(ctx context.Context, t *transfer.Transfer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t.ID = uuid.New().String()
	r.transfers[t.ID] = t
	return nil
}

func (r *TransferRepository) GetByID(ctx context.Context, id string) (*transfer.Transfer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.transfers[id]
	if !ok {
		return nil, transfer.ErrTransferNotFound
	}
	return t, nil
}

func (r *TransferRepository) ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*transfer.Transfer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []*transfer.Transfer
	for _, t := range r.transfers {
		if t.FromAccountID == accountID || t.ToAccountID == accountID {
			all = append(all, t)
		}
	}

	if offset >= len(all) {
		return nil, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (r *TransferRepository) CountByAccountID(ctx context.Context, accountID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, t := range r.transfers {
		if t.FromAccountID == accountID || t.ToAccountID == accountID {
			count++
		}
	}
	return count, nil
}

// Seed - test uchun transfer qo'shish
func (r *TransferRepository) Seed(t *transfer.Transfer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	t.Amount.Currency = shared.Currency(t.Amount.Currency)
	r.transfers[t.ID] = t
}
