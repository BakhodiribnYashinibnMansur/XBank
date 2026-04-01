package postgres

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/transfer"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TransferProjectionRepository - manages the CQRS read projection for transfers.
// Denormalized: includes account numbers and user IDs to avoid JOINs.
type TransferProjectionRepository struct {
	pool *pgxpool.Pool
}

func NewTransferProjectionRepository(pool *pgxpool.Pool) *TransferProjectionRepository {
	return &TransferProjectionRepository{pool: pool}
}

// Upsert - create or update transfer projection after events
func (r *TransferProjectionRepository) Upsert(ctx context.Context, t *transfer.Transfer, fromAccountNumber, toAccountNumber, fromUserID, toUserID string) error {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)

	_, err := db.Exec(ctx,
		`INSERT INTO transfer_projections
			(id, from_account_id, from_account_number, from_user_id,
			 to_account_id, to_account_number, to_user_id,
			 amount, currency, status, description, failure_reason, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		 ON CONFLICT (id)
		 DO UPDATE SET status = EXCLUDED.status,
		               failure_reason = EXCLUDED.failure_reason,
		               completed_at = CASE WHEN EXCLUDED.status IN ('COMPLETED','FAILED') THEN NOW() ELSE NULL END`,
		t.ID, t.FromAccountID, fromAccountNumber, fromUserID,
		t.ToAccountID, toAccountNumber, toUserID,
		t.Amount.Amount, t.Amount.Currency, t.Status,
		t.Description, t.FailureReason, t.CreatedAt,
	)
	metrics.ObserveQuery("TransferProjection.Upsert", start, err)
	return err
}

// TransferSummary - denormalized read model for user-facing transfer list
type TransferSummary struct {
	ID                 string
	FromAccountID      string
	FromAccountNumber  string
	FromUserID         string
	ToAccountID        string
	ToAccountNumber    string
	ToUserID           string
	Amount             int64
	Currency           string
	Status             string
	Description        string
	FailureReason      string
	CreatedAt          time.Time
	CompletedAt        *time.Time
}

// ListByUserID - all transfers where user is sender or receiver
func (r *TransferProjectionRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*TransferSummary, int64, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)

	var total int64
	if err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM transfer_projections WHERE from_user_id = $1 OR to_user_id = $1`, userID,
	).Scan(&total); err != nil {
		metrics.ObserveQuery("TransferProjection.ListByUserID.Count", start, err)
		return nil, 0, err
	}

	rows, err := db.Query(ctx,
		`SELECT id, from_account_id, from_account_number, from_user_id,
		        to_account_id, to_account_number, to_user_id,
		        amount, currency, status, description, failure_reason,
		        created_at, completed_at
		 FROM transfer_projections
		 WHERE from_user_id = $1 OR to_user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		metrics.ObserveQuery("TransferProjection.ListByUserID", start, err)
		return nil, 0, err
	}
	defer rows.Close()

	var summaries []*TransferSummary
	for rows.Next() {
		s := &TransferSummary{}
		if err := rows.Scan(&s.ID, &s.FromAccountID, &s.FromAccountNumber, &s.FromUserID,
			&s.ToAccountID, &s.ToAccountNumber, &s.ToUserID,
			&s.Amount, &s.Currency, &s.Status, &s.Description, &s.FailureReason,
			&s.CreatedAt, &s.CompletedAt); err != nil {
			return nil, 0, err
		}
		summaries = append(summaries, s)
	}

	metrics.ObserveQuery("TransferProjection.ListByUserID", start, nil)
	return summaries, total, nil
}

// ListByAccountID - transfers for a specific account
func (r *TransferProjectionRepository) ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*TransferSummary, int64, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)

	var total int64
	if err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM transfer_projections WHERE from_account_id = $1 OR to_account_id = $1`, accountID,
	).Scan(&total); err != nil {
		metrics.ObserveQuery("TransferProjection.ListByAccountID.Count", start, err)
		return nil, 0, err
	}

	rows, err := db.Query(ctx,
		`SELECT id, from_account_id, from_account_number, from_user_id,
		        to_account_id, to_account_number, to_user_id,
		        amount, currency, status, description, failure_reason,
		        created_at, completed_at
		 FROM transfer_projections
		 WHERE from_account_id = $1 OR to_account_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		accountID, limit, offset,
	)
	if err != nil {
		metrics.ObserveQuery("TransferProjection.ListByAccountID", start, err)
		return nil, 0, err
	}
	defer rows.Close()

	var summaries []*TransferSummary
	for rows.Next() {
		s := &TransferSummary{}
		if err := rows.Scan(&s.ID, &s.FromAccountID, &s.FromAccountNumber, &s.FromUserID,
			&s.ToAccountID, &s.ToAccountNumber, &s.ToUserID,
			&s.Amount, &s.Currency, &s.Status, &s.Description, &s.FailureReason,
			&s.CreatedAt, &s.CompletedAt); err != nil {
			return nil, 0, err
		}
		summaries = append(summaries, s)
	}

	metrics.ObserveQuery("TransferProjection.ListByAccountID", start, nil)
	return summaries, total, nil
}
