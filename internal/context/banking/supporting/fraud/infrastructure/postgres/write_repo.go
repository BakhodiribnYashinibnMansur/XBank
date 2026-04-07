package postgres

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/fraud"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FraudRepository struct {
	pool *pgxpool.Pool
}

func NewFraudRepository(pool *pgxpool.Pool) *FraudRepository {
	return &FraudRepository{pool: pool}
}

func (r *FraudRepository) Create(ctx context.Context, ch *fraud.Check) error {
	db := ExtractDBTX(ctx, r.pool)
	return db.QueryRow(ctx,
		`INSERT INTO fraud_checks (transfer_id, user_id, risk_score, risk_level, action, flags, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		ch.TransferID, ch.UserID, ch.RiskScore, ch.RiskLevel, ch.Action, ch.Flags, ch.CreatedAt,
	).Scan(&ch.ID)
}

func (r *FraudRepository) GetByTransferID(ctx context.Context, transferID string) (*fraud.Check, error) {
	db := ExtractDBTX(ctx, r.pool)
	ch := &fraud.Check{}
	err := db.QueryRow(ctx,
		`SELECT id, transfer_id, user_id, risk_score, risk_level, action, flags, reviewed_by, review_comment, created_at
		 FROM fraud_checks WHERE transfer_id = $1`, transferID,
	).Scan(&ch.ID, &ch.TransferID, &ch.UserID, &ch.RiskScore, &ch.RiskLevel, &ch.Action, &ch.Flags, &ch.ReviewedBy, &ch.ReviewComment, &ch.CreatedAt)
	if err != nil {
		return nil, fraud.ErrFraudDetected
	}
	return ch, nil
}

func (r *FraudRepository) ListFlagged(ctx context.Context, limit, offset int) ([]*fraud.Check, error) {
	db := ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, transfer_id, user_id, risk_score, risk_level, action, flags, reviewed_by, review_comment, created_at
		 FROM fraud_checks WHERE action IN ('FLAG', 'BLOCK') ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*fraud.Check
	for rows.Next() {
		ch := &fraud.Check{}
		if err := rows.Scan(&ch.ID, &ch.TransferID, &ch.UserID, &ch.RiskScore, &ch.RiskLevel, &ch.Action, &ch.Flags, &ch.ReviewedBy, &ch.ReviewComment, &ch.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, ch)
	}
	return items, nil
}

func (r *FraudRepository) CountFlagged(ctx context.Context) (int64, error) {
	db := ExtractDBTX(ctx, r.pool)
	var count int64
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM fraud_checks WHERE action IN ('FLAG', 'BLOCK')`).Scan(&count)
	return count, err
}

func (r *FraudRepository) Update(ctx context.Context, ch *fraud.Check) error {
	db := ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx,
		`UPDATE fraud_checks SET reviewed_by = $1, review_comment = $2 WHERE id = $3`,
		ch.ReviewedBy, ch.ReviewComment, ch.ID)
	return err
}
