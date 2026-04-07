package postgres

import (
	"context"
	"fmt"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/exchange/domain"
	sharedpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WriteRepo struct {
	pool *pgxpool.Pool
}

func NewWriteRepo(pool *pgxpool.Pool) *WriteRepo {
	return &WriteRepo{pool: pool}
}

func (r *WriteRepo) Upsert(ctx context.Context, rate *domain.Rate) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rate.UpdatedAt = time.Now()
	err := db.QueryRow(ctx,
		`INSERT INTO exchange_rates (from_currency, to_currency, buy_rate, sell_rate, updated_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (from_currency, to_currency) DO UPDATE
		 SET buy_rate = $3, sell_rate = $4, updated_at = $5
		 RETURNING id`,
		rate.FromCurrency, rate.ToCurrency, rate.BuyRate, rate.SellRate, rate.UpdatedAt,
	).Scan(&rate.ID)
	metrics.ObserveQuery("ExchangeRepo.Upsert", start, err)
	if err != nil {
		return fmt.Errorf("exchange_repo: upsert: %w", err)
	}
	return nil
}

func (r *WriteRepo) GetRate(ctx context.Context, from, to string) (*domain.Rate, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rate := &domain.Rate{}
	err := db.QueryRow(ctx,
		`SELECT id, from_currency, to_currency, buy_rate, sell_rate, updated_at
		 FROM exchange_rates WHERE from_currency = $1 AND to_currency = $2`,
		from, to,
	).Scan(&rate.ID, &rate.FromCurrency, &rate.ToCurrency, &rate.BuyRate, &rate.SellRate, &rate.UpdatedAt)
	metrics.ObserveQuery("ExchangeRepo.GetRate", start, err)
	if err != nil {
		return nil, domain.ErrRateNotFound
	}
	return rate, nil
}

func (r *WriteRepo) ListAll(ctx context.Context) ([]*domain.Rate, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, from_currency, to_currency, buy_rate, sell_rate, updated_at
		 FROM exchange_rates ORDER BY from_currency, to_currency`)
	if err != nil {
		metrics.ObserveQuery("ExchangeRepo.ListAll", start, err)
		return nil, fmt.Errorf("exchange_repo: list: %w", err)
	}
	defer rows.Close()

	var rates []*domain.Rate
	for rows.Next() {
		rate := &domain.Rate{}
		if err := rows.Scan(&rate.ID, &rate.FromCurrency, &rate.ToCurrency, &rate.BuyRate, &rate.SellRate, &rate.UpdatedAt); err != nil {
			metrics.ObserveQuery("ExchangeRepo.ListAll", start, err)
			return nil, fmt.Errorf("exchange_repo: list scan: %w", err)
		}
		rates = append(rates, rate)
	}
	metrics.ObserveQuery("ExchangeRepo.ListAll", start, nil)
	return rates, nil
}
