package command

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/fraud"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
)

type Service struct {
	repo fraud.Repository
}

func NewService(repo fraud.Repository) *Service {
	return &Service{repo: repo}
}

// Evaluate - run fraud/AML check on a transfer
func (s *Service) Evaluate(ctx context.Context, transferID, userID string, amount int64, flags []string) (_ *fraud.Check, err error) {
	defer metrics.ObserveService("FraudService", "Evaluate", time.Now(), &err)

	check := fraud.NewCheck(transferID, userID, amount, flags)
	if err := s.repo.Create(ctx, check); err != nil {
		return nil, err
	}
	return check, nil
}

// ShouldBlock - quick check if a transfer should be blocked
func (s *Service) ShouldBlock(ctx context.Context, transferID, userID string, amount int64, flags []string) (bool, *fraud.Check) {
	check := fraud.NewCheck(transferID, userID, amount, flags)
	s.repo.Create(ctx, check) // best-effort save
	return check.Action == fraud.ActionBlock, check
}

func (s *Service) GetByTransferID(ctx context.Context, transferID string) (_ *fraud.Check, err error) {
	defer metrics.ObserveService("FraudService", "GetByTransferID", time.Now(), &err)
	return s.repo.GetByTransferID(ctx, transferID)
}

func (s *Service) ListFlagged(ctx context.Context, limit, offset int) (_ []*fraud.Check, _ int64, err error) {
	defer metrics.ObserveService("FraudService", "ListFlagged", time.Now(), &err)

	items, err := s.repo.ListFlagged(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.CountFlagged(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
