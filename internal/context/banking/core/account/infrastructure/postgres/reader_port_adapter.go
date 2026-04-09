package postgres

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/contract/ports"
	account "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/domain"
	sharedpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ReaderPortAdapter implements ports.AccountReader.
type ReaderPortAdapter struct {
	pool *pgxpool.Pool
}

func NewReaderPortAdapter(pool *pgxpool.Pool) *ReaderPortAdapter {
	return &ReaderPortAdapter{pool: pool}
}

func (a *ReaderPortAdapter) GetByID(ctx context.Context, id string) (*ports.AccountView, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, a.pool)
	v := &ports.AccountView{}
	err := db.QueryRow(ctx,
		`SELECT id, user_id, account_number, currency, balance, status
		 FROM accounts WHERE id = $1`, id,
	).Scan(&v.ID, &v.UserID, &v.AccountNumber, &v.Currency, &v.Balance, &v.Status)
	metrics.ObserveQuery("AccountReaderPort.GetByID", start, err)
	if err != nil {
		return nil, account.ErrAccountNotFound
	}
	return v, nil
}

func (a *ReaderPortAdapter) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*ports.AccountView, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, a.pool)
	rows, err := db.Query(ctx,
		`SELECT id, user_id, account_number, currency, balance, status
		 FROM accounts WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	metrics.ObserveQuery("AccountReaderPort.ListByUserID", start, err)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*ports.AccountView
	for rows.Next() {
		v := &ports.AccountView{}
		if err := rows.Scan(&v.ID, &v.UserID, &v.AccountNumber, &v.Currency, &v.Balance, &v.Status); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

func (a *ReaderPortAdapter) CountByUserID(ctx context.Context, userID string) (int64, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, a.pool)
	var count int64
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM accounts WHERE user_id = $1`, userID).Scan(&count)
	metrics.ObserveQuery("AccountReaderPort.CountByUserID", start, err)
	return count, err
}
