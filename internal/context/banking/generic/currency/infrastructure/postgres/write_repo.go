package postgres

import (
	"context"
	"fmt"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/generic/currency/domain"
	sharedpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WriteRepo implements currency.Repository using PostgreSQL.
type WriteRepo struct {
	pool *pgxpool.Pool
}

// NewWriteRepo creates a new currency postgres repository.
func NewWriteRepo(pool *pgxpool.Pool) *WriteRepo {
	return &WriteRepo{pool: pool}
}

func (r *WriteRepo) Create(ctx context.Context, c *domain.Currency) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	c.CreatedAt = time.Now()
	c.UpdatedAt = c.CreatedAt

	err := db.QueryRow(ctx,
		`INSERT INTO currencies (code, name, symbol, decimal_places, status, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id`,
		c.Code, c.Name, c.Symbol, c.DecimalPlaces, c.Status, c.CreatedAt, c.UpdatedAt,
	).Scan(&c.ID)
	metrics.ObserveQuery("CurrencyRepo.Create", start, err)
	if err != nil {
		return fmt.Errorf("currency_repo: create: %w", err)
	}
	return nil
}

func (r *WriteRepo) GetByID(ctx context.Context, id string) (*domain.Currency, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	c := &domain.Currency{}
	err := db.QueryRow(ctx,
		`SELECT id, code, name, symbol, decimal_places, status, created_at, updated_at
		 FROM currencies WHERE id = $1`, id,
	).Scan(&c.ID, &c.Code, &c.Name, &c.Symbol, &c.DecimalPlaces, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	metrics.ObserveQuery("CurrencyRepo.GetByID", start, err)
	if err != nil {
		return nil, domain.ErrCurrencyNotFound
	}
	return c, nil
}

func (r *WriteRepo) GetByCode(ctx context.Context, code string) (*domain.Currency, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	c := &domain.Currency{}
	err := db.QueryRow(ctx,
		`SELECT id, code, name, symbol, decimal_places, status, created_at, updated_at
		 FROM currencies WHERE code = $1`, code,
	).Scan(&c.ID, &c.Code, &c.Name, &c.Symbol, &c.DecimalPlaces, &c.Status, &c.CreatedAt, &c.UpdatedAt)
	metrics.ObserveQuery("CurrencyRepo.GetByCode", start, err)
	if err != nil {
		return nil, domain.ErrCurrencyNotFound
	}
	return c, nil
}

func (r *WriteRepo) ListAll(ctx context.Context) ([]*domain.Currency, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, code, name, symbol, decimal_places, status, created_at, updated_at
		 FROM currencies ORDER BY code`)
	if err != nil {
		metrics.ObserveQuery("CurrencyRepo.ListAll", start, err)
		return nil, fmt.Errorf("currency_repo: list: %w", err)
	}
	defer rows.Close()

	var currencies []*domain.Currency
	for rows.Next() {
		c := &domain.Currency{}
		if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.Symbol, &c.DecimalPlaces, &c.Status, &c.CreatedAt, &c.UpdatedAt); err != nil {
			metrics.ObserveQuery("CurrencyRepo.ListAll", start, err)
			return nil, fmt.Errorf("currency_repo: list scan: %w", err)
		}
		currencies = append(currencies, c)
	}
	metrics.ObserveQuery("CurrencyRepo.ListAll", start, nil)
	return currencies, nil
}

func (r *WriteRepo) Update(ctx context.Context, c *domain.Currency) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	c.UpdatedAt = time.Now()

	_, err := db.Exec(ctx,
		`UPDATE currencies SET name = $1, symbol = $2, decimal_places = $3, status = $4, updated_at = $5
		 WHERE id = $6`,
		c.Name, c.Symbol, c.DecimalPlaces, c.Status, c.UpdatedAt, c.ID,
	)
	metrics.ObserveQuery("CurrencyRepo.Update", start, err)
	if err != nil {
		return fmt.Errorf("currency_repo: update: %w", err)
	}
	return nil
}
