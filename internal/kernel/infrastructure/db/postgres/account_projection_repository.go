package postgres

import (
	"context"
	"time"

	account "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AccountProjectionRepository - manages the CQRS read projection.
// Writes go to account_projections (denormalized, read-optimized).
// This is updated synchronously after each event append.
type AccountProjectionRepository struct {
	pool *pgxpool.Pool
}

func NewAccountProjectionRepository(pool *pgxpool.Pool) *AccountProjectionRepository {
	return &AccountProjectionRepository{pool: pool}
}

// Upsert - create or update projection after events are applied.
// Called from the application service after saveAggregate.
func (r *AccountProjectionRepository) Upsert(ctx context.Context, acc *account.Account, userEmail string) error {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)

	_, err := db.Exec(ctx,
		`INSERT INTO account_projections
			(id, user_id, user_email, account_number, balance, currency, status, created_at, updated_at, last_event_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (id)
		 DO UPDATE SET balance = EXCLUDED.balance,
		               status = EXCLUDED.status,
		               updated_at = EXCLUDED.updated_at,
		               last_event_at = EXCLUDED.last_event_at,
		               event_count = account_projections.event_count + 1`,
		acc.ID, acc.UserID, userEmail, acc.AccountNumber,
		acc.Balance.Amount, acc.Balance.Currency, acc.Status,
		acc.CreatedAt, acc.UpdatedAt, acc.UpdatedAt,
	)
	metrics.ObserveQuery("AccountProjection.Upsert", start, err)
	return err
}

// UpdateCounters - increment total_credited or total_debited.
func (r *AccountProjectionRepository) AddCredit(ctx context.Context, accountID string, amount int64) error {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)

	_, err := db.Exec(ctx,
		`UPDATE account_projections SET total_credited = total_credited + $1 WHERE id = $2`,
		amount, accountID,
	)
	metrics.ObserveQuery("AccountProjection.AddCredit", start, err)
	return err
}

func (r *AccountProjectionRepository) AddDebit(ctx context.Context, accountID string, amount int64) error {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)

	_, err := db.Exec(ctx,
		`UPDATE account_projections SET total_debited = total_debited + $1 WHERE id = $2`,
		amount, accountID,
	)
	metrics.ObserveQuery("AccountProjection.AddDebit", start, err)
	return err
}

// AccountSummary - denormalized read model for dashboard display
type AccountSummary struct {
	ID             string
	UserID         string
	UserEmail      string
	AccountNumber  string
	Balance        int64
	Currency       string
	Status         string
	TotalCredited  int64
	TotalDebited   int64
	EventCount     int
	LastEventAt    *time.Time
	CreatedAt      time.Time
}

// GetSummary - fetch enriched account summary from projection
func (r *AccountProjectionRepository) GetSummary(ctx context.Context, accountID string) (*AccountSummary, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)

	s := &AccountSummary{}
	err := db.QueryRow(ctx,
		`SELECT id, user_id, user_email, account_number, balance, currency, status,
		        total_credited, total_debited, event_count, last_event_at, created_at
		 FROM account_projections WHERE id = $1`,
		accountID,
	).Scan(&s.ID, &s.UserID, &s.UserEmail, &s.AccountNumber, &s.Balance, &s.Currency,
		&s.Status, &s.TotalCredited, &s.TotalDebited, &s.EventCount, &s.LastEventAt, &s.CreatedAt)

	metrics.ObserveQuery("AccountProjection.GetSummary", start, err)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// ListByUserID - paginated list from projection (no JOIN needed)
func (r *AccountProjectionRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*AccountSummary, int64, error) {
	start := time.Now()
	db := ExtractDBTX(ctx, r.pool)

	// Count
	var total int64
	if err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM account_projections WHERE user_id = $1`, userID,
	).Scan(&total); err != nil {
		metrics.ObserveQuery("AccountProjection.ListByUserID.Count", start, err)
		return nil, 0, err
	}

	// List
	rows, err := db.Query(ctx,
		`SELECT id, user_id, user_email, account_number, balance, currency, status,
		        total_credited, total_debited, event_count, last_event_at, created_at
		 FROM account_projections
		 WHERE user_id = $1
		 ORDER BY created_at DESC
		 LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		metrics.ObserveQuery("AccountProjection.ListByUserID", start, err)
		return nil, 0, err
	}
	defer rows.Close()

	var summaries []*AccountSummary
	for rows.Next() {
		s := &AccountSummary{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.UserEmail, &s.AccountNumber, &s.Balance, &s.Currency,
			&s.Status, &s.TotalCredited, &s.TotalDebited, &s.EventCount, &s.LastEventAt, &s.CreatedAt); err != nil {
			return nil, 0, err
		}
		summaries = append(summaries, s)
	}

	metrics.ObserveQuery("AccountProjection.ListByUserID", start, nil)
	return summaries, total, nil
}
