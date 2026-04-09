package postgres

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	transfer "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScheduledRepo struct {
	pool *pgxpool.Pool
}

func NewScheduledRepo(pool *pgxpool.Pool) *ScheduledRepo {
	return &ScheduledRepo{pool: pool}
}

func (r *ScheduledRepo) Create(ctx context.Context, st *transfer.ScheduledTransfer) error {
	start := time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO scheduled_transfers (id, user_id, from_account_id, to_account_id, amount, currency, description, status, execute_at, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		st.ID, st.UserID, st.FromAccountID, st.ToAccountID,
		st.Amount.Amount, st.Amount.Currency, st.Description,
		st.Status, st.ExecuteAt, st.CreatedAt,
	)
	metrics.ObserveQuery("ScheduledTransferRepo.Create", start, err)
	return err
}

func (r *ScheduledRepo) GetByID(ctx context.Context, id string) (*transfer.ScheduledTransfer, error) {
	start := time.Now()
	st := &transfer.ScheduledTransfer{}
	var currency string
	var transferID *string

	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, from_account_id, to_account_id, amount, currency, description,
		        status, execute_at, transfer_id, failure_reason, created_at, executed_at
		 FROM scheduled_transfers WHERE id = $1`,
		id,
	).Scan(&st.ID, &st.UserID, &st.FromAccountID, &st.ToAccountID,
		&st.Amount.Amount, &currency, &st.Description,
		&st.Status, &st.ExecuteAt, &transferID, &st.FailureReason,
		&st.CreatedAt, &st.ExecutedAt)

	metrics.ObserveQuery("ScheduledTransferRepo.GetByID", start, err)
	if err != nil {
		return nil, err
	}
	st.Amount.Currency = domain.Currency(currency)
	if transferID != nil {
		st.TransferID = *transferID
	}
	return st, nil
}

func (r *ScheduledRepo) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*transfer.ScheduledTransfer, int64, error) {
	start := time.Now()

	var total int64
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM scheduled_transfers WHERE user_id = $1`, userID,
	).Scan(&total); err != nil {
		metrics.ObserveQuery("ScheduledTransferRepo.ListByUserID.Count", start, err)
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, from_account_id, to_account_id, amount, currency, description,
		        status, execute_at, transfer_id, failure_reason, created_at, executed_at
		 FROM scheduled_transfers WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		metrics.ObserveQuery("ScheduledTransferRepo.ListByUserID", start, err)
		return nil, 0, err
	}
	defer rows.Close()

	var results []*transfer.ScheduledTransfer
	for rows.Next() {
		st := &transfer.ScheduledTransfer{}
		var currency string
		var transferID *string
		if err := rows.Scan(&st.ID, &st.UserID, &st.FromAccountID, &st.ToAccountID,
			&st.Amount.Amount, &currency, &st.Description,
			&st.Status, &st.ExecuteAt, &transferID, &st.FailureReason,
			&st.CreatedAt, &st.ExecutedAt); err != nil {
			return nil, 0, err
		}
		st.Amount.Currency = domain.Currency(currency)
		if transferID != nil {
			st.TransferID = *transferID
		}
		results = append(results, st)
	}

	metrics.ObserveQuery("ScheduledTransferRepo.ListByUserID", start, nil)
	return results, total, nil
}

func (r *ScheduledRepo) Update(ctx context.Context, st *transfer.ScheduledTransfer) error {
	start := time.Now()
	var transferID interface{}
	if st.TransferID != "" {
		transferID = st.TransferID
	}
	_, err := r.pool.Exec(ctx,
		`UPDATE scheduled_transfers
		 SET status = $1, transfer_id = $2, failure_reason = $3, executed_at = $4
		 WHERE id = $5`,
		st.Status, transferID, st.FailureReason, st.ExecutedAt, st.ID,
	)
	metrics.ObserveQuery("ScheduledTransferRepo.Update", start, err)
	return err
}

// FetchDue - get pending transfers whose execute_at has passed.
// Uses FOR UPDATE SKIP LOCKED to allow concurrent workers.
func (r *ScheduledRepo) FetchDue(ctx context.Context, limit int) ([]*transfer.ScheduledTransfer, error) {
	start := time.Now()
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, from_account_id, to_account_id, amount, currency, description,
		        status, execute_at, transfer_id, failure_reason, created_at, executed_at
		 FROM scheduled_transfers
		 WHERE status = 'PENDING' AND execute_at <= NOW()
		 ORDER BY execute_at ASC
		 LIMIT $1
		 FOR UPDATE SKIP LOCKED`,
		limit,
	)
	if err != nil {
		metrics.ObserveQuery("ScheduledTransferRepo.FetchDue", start, err)
		return nil, err
	}
	defer rows.Close()

	var results []*transfer.ScheduledTransfer
	for rows.Next() {
		st := &transfer.ScheduledTransfer{}
		var currency string
		var transferID *string
		if err := rows.Scan(&st.ID, &st.UserID, &st.FromAccountID, &st.ToAccountID,
			&st.Amount.Amount, &currency, &st.Description,
			&st.Status, &st.ExecuteAt, &transferID, &st.FailureReason,
			&st.CreatedAt, &st.ExecutedAt); err != nil {
			return nil, err
		}
		st.Amount.Currency = domain.Currency(currency)
		if transferID != nil {
			st.TransferID = *transferID
		}
		results = append(results, st)
	}

	metrics.ObserveQuery("ScheduledTransferRepo.FetchDue", start, nil)
	return results, nil
}
