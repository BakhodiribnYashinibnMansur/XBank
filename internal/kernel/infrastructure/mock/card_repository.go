package mock

import (
	"context"
	"sync"

	card "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/card/domain"
	"github.com/google/uuid"
)

type CardRepository struct {
	mu    sync.RWMutex
	cards map[string]*card.Card
}

func NewCardRepository() *CardRepository {
	return &CardRepository{cards: make(map[string]*card.Card)}
}

func (r *CardRepository) Create(ctx context.Context, c *card.Card) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c.ID = uuid.New().String()
	r.cards[c.ID] = c
	return nil
}

func (r *CardRepository) GetByID(ctx context.Context, id string) (*card.Card, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.cards[id]
	if !ok {
		return nil, card.ErrCardNotFound
	}
	return c, nil
}

func (r *CardRepository) ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*card.Card, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var all []*card.Card
	for _, c := range r.cards {
		if c.AccountID == accountID {
			all = append(all, c)
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

func (r *CardRepository) CountByAccountID(ctx context.Context, accountID string) (int64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var count int64
	for _, c := range r.cards {
		if c.AccountID == accountID {
			count++
		}
	}
	return count, nil
}

func (r *CardRepository) Update(ctx context.Context, c *card.Card) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cards[c.ID] = c
	return nil
}
