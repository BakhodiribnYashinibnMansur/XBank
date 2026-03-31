package postgres

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/transfer"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransferRepository struct {
	pool *pgxpool.Pool
}

func NewTransferRepository(pool *pgxpool.Pool) *TransferRepository {
	return &TransferRepository{pool: pool}
}

func (r *TransferRepository) Create(ctx context.Context, t *transfer.Transfer) error {
	db := ExtractDBTX(ctx, r.pool)
	query := `
		INSERT INTO transfers (from_account_id, to_account_id, amount, currency, status, description, failure_reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	return db.QueryRow(ctx, query,
		t.FromAccountID, t.ToAccountID, t.Amount.Amount,
		t.Amount.Currency, t.Status, t.Description, t.FailureReason, t.CreatedAt,
	).Scan(&t.ID)
}

func (r *TransferRepository) GetByID(ctx context.Context, id string) (*transfer.Transfer, error) {
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
	if err != nil {
		return nil, transfer.ErrTransferNotFound
	}
	t.Amount.Currency = shared.Currency(currency)
	return t, nil
}

func (r *TransferRepository) ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*transfer.Transfer, error) {
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
		t.Amount.Currency = shared.Currency(currency)
		transfers = append(transfers, t)
	}
	return transfers, nil
}

func (r *TransferRepository) CountByAccountID(ctx context.Context, accountID string) (int64, error) {
	db := ExtractDBTX(ctx, r.pool)
	var count int64
	err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM transfers WHERE from_account_id = $1 OR to_account_id = $1`,
		accountID,
	).Scan(&count)
	return count, err
}
