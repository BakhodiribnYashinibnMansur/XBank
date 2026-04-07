package command

import (
	"context"
	"time"

	domainCard "github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/card"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.uber.org/zap"
)

// HoldService - manages card authorization holds (hold/capture/release)
type HoldService struct {
	holdRepo domainCard.HoldRepository
	cardRepo domainCard.Repository
}

func NewHoldService(holdRepo domainCard.HoldRepository, cardRepo domainCard.Repository) *HoldService {
	return &HoldService{holdRepo: holdRepo, cardRepo: cardRepo}
}

// Hold - create an authorization hold (reserve funds)
func (s *HoldService) Hold(ctx context.Context, cardID, accountID, merchant string, amount int64, currency string) (_ *domainCard.Hold, err error) {
	defer metrics.ObserveService("HoldService", "Hold", time.Now(), &err)

	// Verify card is usable
	c, err := s.cardRepo.GetByID(ctx, cardID)
	if err != nil {
		return nil, domainCard.ErrCardNotFound
	}
	if c.Status != domainCard.StatusActive {
		return nil, apperror.ErrValidation.WithMessage("Card is not active")
	}

	hold, err := domainCard.NewHold(cardID, accountID, merchant, amount, currency)
	if err != nil {
		return nil, apperror.ErrBadRequest.WithMessage(err.Error())
	}

	if err := s.holdRepo.Create(ctx, hold); err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}

	return hold, nil
}

// Capture - settle an authorization hold (debit the account)
func (s *HoldService) Capture(ctx context.Context, holdID string) (_ *domainCard.Hold, err error) {
	defer metrics.ObserveService("HoldService", "Capture", time.Now(), &err)

	hold, err := s.holdRepo.GetByID(ctx, holdID)
	if err != nil {
		return nil, apperror.ErrNotFound.WithMessage("Hold not found")
	}

	if err := hold.Capture(); err != nil {
		return nil, apperror.ErrBadRequest.WithMessage(err.Error())
	}

	if err := s.holdRepo.Update(ctx, hold); err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}

	return hold, nil
}

// Release - cancel an authorization hold (restore funds)
func (s *HoldService) Release(ctx context.Context, holdID string) (_ *domainCard.Hold, err error) {
	defer metrics.ObserveService("HoldService", "Release", time.Now(), &err)

	hold, err := s.holdRepo.GetByID(ctx, holdID)
	if err != nil {
		return nil, apperror.ErrNotFound.WithMessage("Hold not found")
	}

	if err := hold.Release(); err != nil {
		return nil, apperror.ErrBadRequest.WithMessage(err.Error())
	}

	if err := s.holdRepo.Update(ctx, hold); err != nil {
		return nil, apperror.ErrInternal.Wrap(err)
	}

	return hold, nil
}

// ListByCard - list holds for a card
func (s *HoldService) ListByCard(ctx context.Context, cardID string) (_ []*domainCard.Hold, err error) {
	defer metrics.ObserveService("HoldService", "ListByCard", time.Now(), &err)
	return s.holdRepo.ListByCardID(ctx, cardID)
}

// ExpireStale - background worker: expire holds past their TTL
func (s *HoldService) ExpireStale(ctx context.Context, batchSize int) (expired int) {
	holds, err := s.holdRepo.FetchExpired(ctx, batchSize)
	if err != nil {
		logger.Log.Error("Failed to fetch expired holds", zap.Error(err))
		return 0
	}

	for _, h := range holds {
		h.Status = domainCard.HoldStatusExpired
		if err := s.holdRepo.Update(ctx, h); err != nil {
			logger.Log.Error("Failed to expire hold", zap.String("hold_id", h.ID), zap.Error(err))
			continue
		}
		expired++
	}

	if expired > 0 {
		logger.Log.Info("Expired stale holds", zap.Int("count", expired))
	}
	return expired
}
