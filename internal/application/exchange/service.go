package exchange

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/exchange"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
)

type Service struct {
	repo exchange.Repository
}

func NewService(repo exchange.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetRate(ctx context.Context, from, to string) (_ *exchange.Rate, err error) {
	defer metrics.ObserveService("ExchangeService", "GetRate", time.Now(), &err)
	return s.repo.GetRate(ctx, from, to)
}

func (s *Service) ListAll(ctx context.Context) (_ []*exchange.Rate, err error) {
	defer metrics.ObserveService("ExchangeService", "ListAll", time.Now(), &err)
	return s.repo.ListAll(ctx)
}

// Convert - convert amount between currencies
func (s *Service) Convert(ctx context.Context, from, to string, amount int64) (_ int64, _ *exchange.Rate, err error) {
	defer metrics.ObserveService("ExchangeService", "Convert", time.Now(), &err)

	rate, err := s.repo.GetRate(ctx, from, to)
	if err != nil {
		return 0, nil, err
	}
	return rate.Convert(amount), rate, nil
}

// UpsertRate - admin: set/update exchange rate
func (s *Service) UpsertRate(ctx context.Context, from, to string, buyRate, sellRate int64) (_ *exchange.Rate, err error) {
	defer metrics.ObserveService("ExchangeService", "UpsertRate", time.Now(), &err)

	rate := &exchange.Rate{
		FromCurrency: from,
		ToCurrency:   to,
		BuyRate:      buyRate,
		SellRate:     sellRate,
	}
	if err := s.repo.Upsert(ctx, rate); err != nil {
		return nil, err
	}
	return rate, nil
}
