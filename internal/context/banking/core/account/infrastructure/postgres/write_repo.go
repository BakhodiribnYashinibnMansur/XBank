package postgres

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/account"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
	sharedPG "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/postgres"
)

type AccountRepository struct {
	pool *pgxpool.Pool
}

func NewAccountRepository(pool *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{pool: pool}
}

func (r *AccountRepository) Create(ctx context.Context, a *account.Account) error {
	start := time.Now()
	db := sharedPG.ExtractDBTX(ctx, r.pool)
	query := `
		INSERT INTO accounts (user_id, account_number, balance, currency, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	err := db.QueryRow(ctx, query,
		a.UserID, a.AccountNumber, a.Balance.Amount, a.Balance.Currency,
		a.Status, a.CreatedAt, a.UpdatedAt,
	).Scan(&a.ID)
	metrics.ObserveQuery("AccountRepo.Create", start, err)
	return err
}

func (r *AccountRepository) GetByID(ctx context.Context, id string) (*account.Account, error) {
	start := time.Now()
	db := sharedPG.ExtractDBTX(ctx, r.pool)
	query := `
		SELECT id, user_id, account_number, balance, currency, status, created_at, updated_at
		FROM accounts WHERE id = $1`

	a := &account.Account{}
	var currency string
	err := db.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.UserID, &a.AccountNumber,
		&a.Balance.Amount, &currency,
		&a.Status, &a.CreatedAt, &a.UpdatedAt,
	)
	metrics.ObserveQuery("AccountRepo.GetByID", start, err)
	if err != nil {
		return nil, account.ErrAccountNotFound
	}
	a.Balance.Currency = shared.Currency(currency)
	return a, nil
}

func (r *AccountRepository) GetByIDForUpdate(ctx context.Context, id string) (*account.Account, error) {
	start := time.Now()
	db := sharedPG.ExtractDBTX(ctx, r.pool)
	query := `
		SELECT id, user_id, account_number, balance, currency, status, created_at, updated_at
		FROM accounts WHERE id = $1 FOR UPDATE`

	a := &account.Account{}
	var currency string
	err := db.QueryRow(ctx, query, id).Scan(
		&a.ID, &a.UserID, &a.AccountNumber,
		&a.Balance.Amount, &currency,
		&a.Status, &a.CreatedAt, &a.UpdatedAt,
	)
	metrics.ObserveQuery("AccountRepo.GetByIDForUpdate", start, err)
	if err != nil {
		return nil, account.ErrAccountNotFound
	}
	a.Balance.Currency = shared.Currency(currency)
	return a, nil
}

func (r *AccountRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*account.Account, error) {
	start := time.Now()
	db := sharedPG.ExtractDBTX(ctx, r.pool)
	query := `
		SELECT id, user_id, account_number, balance, currency, status, created_at, updated_at
		FROM accounts WHERE user_id = $1 ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		metrics.ObserveQuery("AccountRepo.ListByUserID", start, err)
		return nil, err
	}
	defer rows.Close()

	var accounts []*account.Account
	for rows.Next() {
		a := &account.Account{}
		var currency string
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.AccountNumber,
			&a.Balance.Amount, &currency,
			&a.Status, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		a.Balance.Currency = shared.Currency(currency)
		accounts = append(accounts, a)
	}
	metrics.ObserveQuery("AccountRepo.ListByUserID", start, nil)
	return accounts, nil
}

func (r *AccountRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	start := time.Now()
	db := sharedPG.ExtractDBTX(ctx, r.pool)
	var count int64
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM accounts WHERE user_id = $1`, userID).Scan(&count)
	metrics.ObserveQuery("AccountRepo.CountByUserID", start, err)
	return count, err
}

func (r *AccountRepository) Update(ctx context.Context, a *account.Account) error {
	start := time.Now()
	db := sharedPG.ExtractDBTX(ctx, r.pool)
	query := `
		UPDATE accounts SET balance = $1, status = $2, updated_at = $3
		WHERE id = $4`

	_, err := db.Exec(ctx, query, a.Balance.Amount, a.Status, a.UpdatedAt, a.ID)
	metrics.ObserveQuery("AccountRepo.Update", start, err)
	return err
}
