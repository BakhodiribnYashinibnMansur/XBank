package postgres

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/contract/ports"
	sharedpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ReaderPortAdapter implements ports.LedgerReader.
type ReaderPortAdapter struct {
	pool *pgxpool.Pool
}

func NewReaderPortAdapter(pool *pgxpool.Pool) *ReaderPortAdapter {
	return &ReaderPortAdapter{pool: pool}
}

func (a *ReaderPortAdapter) ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*ports.LedgerEntryView, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, a.pool)
	rows, err := db.Query(ctx,
		`SELECT id, account_id, transfer_id, entry_type, amount, currency, created_at
		 FROM ledger_entries WHERE account_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		accountID, limit, offset,
	)
	metrics.ObserveQuery("LedgerReaderPort.ListByAccountID", start, err)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*ports.LedgerEntryView
	for rows.Next() {
		v := &ports.LedgerEntryView{}
		if err := rows.Scan(&v.ID, &v.AccountID, &v.TransferID, &v.EntryType, &v.Amount, &v.Currency, &v.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

func (a *ReaderPortAdapter) CountByAccountID(ctx context.Context, accountID string) (int64, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, a.pool)
	var count int64
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM ledger_entries WHERE account_id = $1`, accountID).Scan(&count)
	metrics.ObserveQuery("LedgerReaderPort.CountByAccountID", start, err)
	return count, err
}

func (a *ReaderPortAdapter) BalanceByAccountID(ctx context.Context, accountID string) (int64, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, a.pool)
	var balance int64
	err := db.QueryRow(ctx,
		`SELECT COALESCE(SUM(CASE WHEN entry_type = 'CREDIT' THEN amount ELSE -amount END), 0)
		 FROM ledger_entries WHERE account_id = $1`, accountID,
	).Scan(&balance)
	metrics.ObserveQuery("LedgerReaderPort.BalanceByAccountID", start, err)
	return balance, err
}
