package query

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/ledger/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
)

// ListResult contains paginated ledger entries.
type ListResult struct {
	Items   []*domain.Entry `json:"items"`
	Total   int64           `json:"total"`
	Balance int64           `json:"balance"`
}

// ListHandler retrieves ledger entries for an account.
type ListHandler struct {
	repo domain.Repository
}

func NewListHandler(repo domain.Repository) *ListHandler {
	return &ListHandler{repo: repo}
}

func (h *ListHandler) Handle(ctx context.Context, accountID string, limit, offset int) (_ *ListResult, err error) {
	defer metrics.ObserveService("LedgerService", "ListByAccount", time.Now(), &err)

	items, err := h.repo.ListByAccountID(ctx, accountID, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := h.repo.CountByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	balance, err := h.repo.BalanceByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	return &ListResult{Items: items, Total: total, Balance: balance}, nil
}
