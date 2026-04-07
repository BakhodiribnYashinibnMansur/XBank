package postgres

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	transfer "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransferRepository struct {
	pool *pgxpool.Pool
}

func NewTransferRepository(pool *pgxpool.Pool) *TransferRepository {
	return &TransferRepository{pool: pool}
}

func (r *TransferRepository) Create(ctx context.Context, t *transfer.Transfer) error {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
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

func (r *TransferRepository) GetByID(ctx context.Context, id string) (*transfer.Transfer, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
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

func (r *TransferRepository) ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*transfer.Transfer, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
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

func (r *TransferRepository) Update(ctx context.Context, t *transfer.Transfer) error {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx,
		`UPDATE transfers SET status = $1, failure_reason = $2 WHERE id = $3`,
		t.Status, t.FailureReason, t.ID,
	)
	metrics.ObserveQuery("TransferRepo.Update", start, err)
	return err
}

func (r *TransferRepository) CountByAccountID(ctx context.Context, accountID string) (int64, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)
	var count int64
	err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM transfers WHERE from_account_id = $1 OR to_account_id = $1`,
		accountID,
	).Scan(&count)
	metrics.ObserveQuery("TransferRepo.CountByAccountID", start, err)
	return count, err
}
