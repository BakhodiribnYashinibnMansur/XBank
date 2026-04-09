package command

import (
	"context"
	"time"

	currency "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/generic/currency/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
)

// Service provides currency catalog operations.
type Service struct {
	repo currency.Repository
}

// NewService creates a new currency service.
func NewService(repo currency.Repository) *Service {
	return &Service{repo: repo}
}

// Create adds a new currency to the catalog.
func (s *Service) Create(ctx context.Context, code, name, symbol string, decimalPlaces int) (_ *currency.Currency, err error) {
	defer metrics.ObserveService("CurrencyService", "Create", time.Now(), &err)

	cur, err := currency.NewCurrency(code, name, symbol, decimalPlaces)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, cur); err != nil {
		return nil, err
	}
	return cur, nil
}

// GetByID returns a currency by its ID.
func (s *Service) GetByID(ctx context.Context, id string) (_ *currency.Currency, err error) {
	defer metrics.ObserveService("CurrencyService", "GetByID", time.Now(), &err)
	return s.repo.GetByID(ctx, id)
}

// GetByCode returns a currency by its ISO code.
func (s *Service) GetByCode(ctx context.Context, code string) (_ *currency.Currency, err error) {
	defer metrics.ObserveService("CurrencyService", "GetByCode", time.Now(), &err)
	return s.repo.GetByCode(ctx, code)
}

// ListAll returns all currencies in the catalog.
func (s *Service) ListAll(ctx context.Context) (_ []*currency.Currency, err error) {
	defer metrics.ObserveService("CurrencyService", "ListAll", time.Now(), &err)
	return s.repo.ListAll(ctx)
}

// Update modifies a currency's details.
func (s *Service) Update(ctx context.Context, id, name, symbol string, decimalPlaces int) (_ *currency.Currency, err error) {
	defer metrics.ObserveService("CurrencyService", "Update", time.Now(), &err)

	cur, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	cur.Name = name
	cur.Symbol = symbol
	cur.DecimalPlaces = decimalPlaces

	if err := s.repo.Update(ctx, cur); err != nil {
		return nil, err
	}
	return cur, nil
}

// ToggleStatus activates or deactivates a currency.
func (s *Service) ToggleStatus(ctx context.Context, id string, active bool) (_ *currency.Currency, err error) {
	defer metrics.ObserveService("CurrencyService", "ToggleStatus", time.Now(), &err)

	cur, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if active {
		cur.Activate()
	} else {
		cur.Deactivate()
	}

	if err := s.repo.Update(ctx, cur); err != nil {
		return nil, err
	}
	return cur, nil
}
