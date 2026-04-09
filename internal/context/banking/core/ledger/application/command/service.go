package command

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/ledger/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
)

// Service provides write operations for the ledger.
type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// RecordTransfer creates a debit+credit pair for a completed transfer.
func (s *Service) RecordTransfer(ctx context.Context, transferID, fromAccountID, toAccountID string, amount int64, currency string) (err error) {
	defer metrics.ObserveService("LedgerService", "RecordTransfer", time.Now(), &err)

	debit, credit := domain.CreatePair(transferID, fromAccountID, toAccountID, amount, currency)
	return s.repo.CreatePair(ctx, debit, credit)
}
