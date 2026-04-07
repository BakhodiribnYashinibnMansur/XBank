package command

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	transfer "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.uber.org/zap"
)

// ScheduledService - manages scheduled (future) transfers.
// Uses the main transfer Service.Send() for actual execution.
type ScheduledService struct {
	schedRepo       transfer.ScheduledTransferRepository
	transferService *Service // the main transfer service
}

func NewScheduledService(
	schedRepo transfer.ScheduledTransferRepository,
	transferService *Service,
) *ScheduledService {
	return &ScheduledService{
		schedRepo:       schedRepo,
		transferService: transferService,
	}
}

// Schedule - create a new scheduled transfer for future execution
func (s *ScheduledService) Schedule(ctx context.Context, userID, fromAccountID, toAccountID string, amount int64, currency domain.Currency, description string, executeAt time.Time) (_ *transfer.ScheduledTransfer, err error) {
	defer metrics.ObserveService("ScheduledTransferService", "Schedule", time.Now(), &err)

	money, err := domain.NewMoney(amount, currency)
	if err != nil {
		return nil, err
	}

	st, err := transfer.NewScheduledTransfer(userID, fromAccountID, toAccountID, money, description, executeAt)
	if err != nil {
		return nil, apperror.ErrBadRequest.WithMessage(err.Error())
	}

	if err := s.schedRepo.Create(ctx, st); err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}

	return st, nil
}

// Cancel - cancel a pending scheduled transfer
func (s *ScheduledService) Cancel(ctx context.Context, id, userID string) (err error) {
	defer metrics.ObserveService("ScheduledTransferService", "Cancel", time.Now(), &err)

	st, err := s.schedRepo.GetByID(ctx, id)
	if err != nil {
		return apperror.ErrNotFound.WithMessage("Scheduled transfer not found")
	}
	if st.UserID != userID {
		return apperror.ErrForbidden
	}

	if err := st.Cancel(); err != nil {
		return apperror.ErrBadRequest.WithMessage(err.Error())
	}

	return s.schedRepo.Update(ctx, st)
}

// GetByID - get a scheduled transfer
func (s *ScheduledService) GetByID(ctx context.Context, id string) (_ *transfer.ScheduledTransfer, err error) {
	defer metrics.ObserveService("ScheduledTransferService", "GetByID", time.Now(), &err)

	st, err := s.schedRepo.GetByID(ctx, id)
	if err != nil {
		return nil, apperror.ErrNotFound.WithMessage("Scheduled transfer not found")
	}
	return st, nil
}

// ListByUserID - list scheduled transfers for a user
func (s *ScheduledService) ListByUserID(ctx context.Context, userID string, limit, offset int) (_ []*transfer.ScheduledTransfer, _ int64, err error) {
	defer metrics.ObserveService("ScheduledTransferService", "ListByUserID", time.Now(), &err)
	return s.schedRepo.ListByUserID(ctx, userID, limit, offset)
}

// ExecuteDue - background worker: fetch and execute due transfers.
// Returns counts of executed and failed transfers.
func (s *ScheduledService) ExecuteDue(ctx context.Context, batchSize int) (executed, failed int) {
	entries, err := s.schedRepo.FetchDue(ctx, batchSize)
	if err != nil {
		logger.Log.Error("Failed to fetch due scheduled transfers", zap.Error(err))
		return 0, 0
	}

	for _, st := range entries {
		tr, err := s.transferService.Send(ctx,
			st.FromAccountID, st.ToAccountID,
			st.Amount.Amount, st.Amount.Currency,
			st.Description,
		)
		if err != nil {
			st.MarkFailed(err.Error())
			s.schedRepo.Update(ctx, st)
			failed++
			logger.Log.Warn("Scheduled transfer failed",
				zap.String("scheduled_id", st.ID),
				zap.Error(err),
			)
		} else {
			st.MarkExecuted(tr.ID)
			s.schedRepo.Update(ctx, st)
			executed++
		}
	}

	if executed+failed > 0 {
		logger.Log.Info("Scheduled transfers processed",
			zap.Int("executed", executed),
			zap.Int("failed", failed),
		)
	}

	return executed, failed
}
