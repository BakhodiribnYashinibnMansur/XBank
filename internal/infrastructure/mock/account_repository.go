package mock

import (
	"context"
	"sync"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/account"
	"github.com/google/uuid"
)

type AccountRepository struct {
	mu       sync.RWMutex
	accounts map[string]*account.Account
}

func NewAccountRepository() *AccountRepository {
	return &AccountRepository{accounts: make(map[string]*account.Account)}
}

func (r *AccountRepository) Create(ctx context.Context, a *account.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	a.ID = uuid.New().String()
	r.accounts[a.ID] = a
	return nil
}

func (r *AccountRepository) GetByID(ctx context.Context, id string) (*account.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	a, ok := r.accounts[id]
	if !ok {
		return nil, account.ErrAccountNotFound
	}
	return a, nil
}

func (r *AccountRepository) GetByIDForUpdate(ctx context.Context, id string) (*account.Account, error) {
	return r.GetByID(ctx, id)
}

func (r *AccountRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*account.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []*account.Account
	for _, a := range r.accounts {
		if a.UserID == userID {
			all = append(all, a)
		}
	}

	// Apply offset and limit
	if offset >= len(all) {
		return nil, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (r *AccountRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int64
	for _, a := range r.accounts {
		if a.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (r *AccountRepository) Update(ctx context.Context, a *account.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.accounts[a.ID] = a
	return nil
}
