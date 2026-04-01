package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/ledger"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LedgerRepository struct {
	pool *pgxpool.Pool
}

func NewLedgerRepository(pool *pgxpool.Pool) *LedgerRepository {
	return &LedgerRepository{pool: pool}
}

func (r *LedgerRepository) CreatePair(ctx context.Context, debit, credit *ledger.Entry) error {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)

	err := db.QueryRow(ctx,
		`INSERT INTO ledger_entries (account_id, transfer_id, entry_type, amount, currency, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		debit.AccountID, debit.TransferID, debit.EntryType, debit.Amount, debit.Currency, debit.CreatedAt,
	).Scan(&debit.ID)
	if err != nil {
		metrics.ObserveQuery("LedgerRepo.CreatePair", start, err)
		return fmt.Errorf("ledger_repo: create debit: %w", err)
	}

	err = db.QueryRow(ctx,
		`INSERT INTO ledger_entries (account_id, transfer_id, entry_type, amount, currency, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		credit.AccountID, credit.TransferID, credit.EntryType, credit.Amount, credit.Currency, credit.CreatedAt,
	).Scan(&credit.ID)
	metrics.ObserveQuery("LedgerRepo.CreatePair", start, err)
	if err != nil {
		return fmt.Errorf("ledger_repo: create credit: %w", err)
	}
	return nil
}

func (r *LedgerRepository) ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*ledger.Entry, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, account_id, transfer_id, entry_type, amount, currency, created_at
		 FROM ledger_entries WHERE account_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		accountID, limit, offset,
	)
	if err != nil {
		metrics.ObserveQuery("LedgerRepo.ListByAccountID", start, err)
		return nil, fmt.Errorf("ledger_repo: list: %w", err)
	}
	defer rows.Close()

	var entries []*ledger.Entry
	for rows.Next() {
		e := &ledger.Entry{}
		if err := rows.Scan(&e.ID, &e.AccountID, &e.TransferID, &e.EntryType, &e.Amount, &e.Currency, &e.CreatedAt); err != nil {
			metrics.ObserveQuery("LedgerRepo.ListByAccountID", start, err)
			return nil, fmt.Errorf("ledger_repo: list scan: %w", err)
		}
		entries = append(entries, e)
	}
	metrics.ObserveQuery("LedgerRepo.ListByAccountID", start, nil)
	return entries, nil
}

func (r *LedgerRepository) CountByAccountID(ctx context.Context, accountID string) (int64, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	var count int64
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM ledger_entries WHERE account_id = $1`, accountID).Scan(&count)
	metrics.ObserveQuery("LedgerRepo.CountByAccountID", start, err)
	return count, err
}

func (r *LedgerRepository) BalanceByAccountID(ctx context.Context, accountID string) (int64, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	var balance int64
	err := db.QueryRow(ctx,
		`SELECT COALESCE(
			SUM(CASE WHEN entry_type = 'CREDIT' THEN amount ELSE -amount END), 0
		 ) FROM ledger_entries WHERE account_id = $1`,
		accountID,
	).Scan(&balance)
	metrics.ObserveQuery("LedgerRepo.BalanceByAccountID", start, err)
	return balance, err
}
