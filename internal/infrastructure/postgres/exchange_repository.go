package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/exchange"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExchangeRepository struct {
	pool *pgxpool.Pool
}

func NewExchangeRepository(pool *pgxpool.Pool) *ExchangeRepository {
	return &ExchangeRepository{pool: pool}
}

func (r *ExchangeRepository) Upsert(ctx context.Context, rate *exchange.Rate) error {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
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

func (r *ExchangeRepository) GetRate(ctx context.Context, from, to string) (*exchange.Rate, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	rate := &exchange.Rate{}
	err := db.QueryRow(ctx,
		`SELECT id, from_currency, to_currency, buy_rate, sell_rate, updated_at
		 FROM exchange_rates WHERE from_currency = $1 AND to_currency = $2`,
		from, to,
	).Scan(&rate.ID, &rate.FromCurrency, &rate.ToCurrency, &rate.BuyRate, &rate.SellRate, &rate.UpdatedAt)
	metrics.ObserveQuery("ExchangeRepo.GetRate", start, err)
	if err != nil {
		return nil, exchange.ErrRateNotFound
	}
	return rate, nil
}

func (r *ExchangeRepository) ListAll(ctx context.Context) ([]*exchange.Rate, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, from_currency, to_currency, buy_rate, sell_rate, updated_at
		 FROM exchange_rates ORDER BY from_currency, to_currency`)
	if err != nil {
		metrics.ObserveQuery("ExchangeRepo.ListAll", start, err)
		return nil, fmt.Errorf("exchange_repo: list: %w", err)
	}
	defer rows.Close()

	var rates []*exchange.Rate
	for rows.Next() {
		rate := &exchange.Rate{}
		if err := rows.Scan(&rate.ID, &rate.FromCurrency, &rate.ToCurrency, &rate.BuyRate, &rate.SellRate, &rate.UpdatedAt); err != nil {
			metrics.ObserveQuery("ExchangeRepo.ListAll", start, err)
			return nil, fmt.Errorf("exchange_repo: list scan: %w", err)
		}
		rates = append(rates, rate)
	}
	metrics.ObserveQuery("ExchangeRepo.ListAll", start, nil)
	return rates, nil
}
