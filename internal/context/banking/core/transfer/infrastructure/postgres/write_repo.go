package postgres

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	transfer "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/domain"
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

func (r *WriteRepo) Create(ctx context.Context, t *transfer.Transfer) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	query := `
		INSERT INTO transfers (id, from_account_id, to_account_id, amount, currency, status, description, failure_reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := db.Exec(ctx, query,
		t.ID, t.FromAccountID, t.ToAccountID, t.Amount.Amount,
		t.Amount.Currency, t.Status, t.Description, t.FailureReason, t.CreatedAt,
	)
	metrics.ObserveQuery("TransferRepo.Create", start, err)
	return err
}

func (r *WriteRepo) GetByID(ctx context.Context, id string) (*transfer.Transfer, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	query := `
		SELECT id, from_account_id, to_account_id, amount, currency,
		       status, description, failure_reason, created_at
		FROM transfers WHERE id = $1`

	t := &transfer.Transfer{}
	var currency string
	err := db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.FromAccountID, &t.ToAccountID,
		&t.Amount.Amount, &currency,
		&t.Status, &t.Description, &t.FailureReason, &t.CreatedAt,
	)
	metrics.ObserveQuery("TransferRepo.GetByID", start, err)
	if err != nil {
		return nil, transfer.ErrTransferNotFound
	}
	t.Amount.Currency = domain.Currency(currency)
	return t, nil
}

func (r *WriteRepo) ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*transfer.Transfer, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	query := `
		SELECT id, from_account_id, to_account_id, amount, currency,
		       status, description, failure_reason, created_at
		FROM transfers
		WHERE from_account_id = $1 OR to_account_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := db.Query(ctx, query, accountID, limit, offset)
	if err != nil {
		metrics.ObserveQuery("TransferRepo.ListByAccountID", start, err)
		return nil, err
	}
	defer rows.Close()

	var transfers []*transfer.Transfer
	for rows.Next() {
		t := &transfer.Transfer{}
		var currency string
		if err := rows.Scan(
			&t.ID, &t.FromAccountID, &t.ToAccountID,
			&t.Amount.Amount, &currency,
			&t.Status, &t.Description, &t.FailureReason, &t.CreatedAt,
		); err != nil {
			return nil, err
		}
		t.Amount.Currency = domain.Currency(currency)
		transfers = append(transfers, t)
	}
	metrics.ObserveQuery("TransferRepo.ListByAccountID", start, nil)
	return transfers, nil
}

func (r *WriteRepo) Update(ctx context.Context, t *transfer.Transfer) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx,
		`UPDATE transfers SET status = $1, failure_reason = $2 WHERE id = $3`,
		t.Status, t.FailureReason, t.ID,
	)
	metrics.ObserveQuery("TransferRepo.Update", start, err)
	return err
}

func (r *WriteRepo) CountByAccountID(ctx context.Context, accountID string) (int64, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	var count int64
	err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM transfers WHERE from_account_id = $1 OR to_account_id = $1`,
		accountID,
	).Scan(&count)
	metrics.ObserveQuery("TransferRepo.CountByAccountID", start, err)
	return count, err
}
