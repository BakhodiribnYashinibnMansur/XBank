package postgres

import (
	"context"
	"fmt"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/fraud/domain"
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

func (r *WriteRepo) Create(ctx context.Context, ch *domain.Check) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	err := db.QueryRow(ctx,
		`INSERT INTO fraud_checks (transfer_id, user_id, risk_score, risk_level, action, flags, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		ch.TransferID, ch.UserID, ch.RiskScore, ch.RiskLevel, ch.Action, ch.Flags, ch.CreatedAt,
	).Scan(&ch.ID)
	metrics.ObserveQuery("FraudRepo.Create", start, err)
	if err != nil {
		return fmt.Errorf("fraud_repo: create: %w", err)
	}
	return nil
}

func (r *WriteRepo) GetByTransferID(ctx context.Context, transferID string) (*domain.Check, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	ch := &domain.Check{}
	err := db.QueryRow(ctx,
		`SELECT id, transfer_id, user_id, risk_score, risk_level, action, flags, reviewed_by, review_comment, created_at
		 FROM fraud_checks WHERE transfer_id = $1`, transferID,
	).Scan(&ch.ID, &ch.TransferID, &ch.UserID, &ch.RiskScore, &ch.RiskLevel, &ch.Action, &ch.Flags, &ch.ReviewedBy, &ch.ReviewComment, &ch.CreatedAt)
	metrics.ObserveQuery("FraudRepo.GetByTransferID", start, err)
	if err != nil {
		return nil, domain.ErrFraudDetected
	}
	return ch, nil
}

func (r *WriteRepo) ListFlagged(ctx context.Context, limit, offset int) ([]*domain.Check, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, transfer_id, user_id, risk_score, risk_level, action, flags, reviewed_by, review_comment, created_at
		 FROM fraud_checks WHERE action IN ('FLAG', 'BLOCK') ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		metrics.ObserveQuery("FraudRepo.ListFlagged", start, err)
		return nil, fmt.Errorf("fraud_repo: list_flagged: %w", err)
	}
	defer rows.Close()

	var items []*domain.Check
	for rows.Next() {
		ch := &domain.Check{}
		if err := rows.Scan(&ch.ID, &ch.TransferID, &ch.UserID, &ch.RiskScore, &ch.RiskLevel, &ch.Action, &ch.Flags, &ch.ReviewedBy, &ch.ReviewComment, &ch.CreatedAt); err != nil {
			metrics.ObserveQuery("FraudRepo.ListFlagged", start, err)
			return nil, fmt.Errorf("fraud_repo: list_flagged scan: %w", err)
		}
		items = append(items, ch)
	}
	metrics.ObserveQuery("FraudRepo.ListFlagged", start, nil)
	return items, nil
}

func (r *WriteRepo) CountFlagged(ctx context.Context) (int64, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	var count int64
	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM fraud_checks WHERE action IN ('FLAG', 'BLOCK')`).Scan(&count)
	metrics.ObserveQuery("FraudRepo.CountFlagged", start, err)
	return count, err
}

func (r *WriteRepo) Update(ctx context.Context, ch *domain.Check) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx,
		`UPDATE fraud_checks SET reviewed_by = $1, review_comment = $2 WHERE id = $3`,
		ch.ReviewedBy, ch.ReviewComment, ch.ID)
	metrics.ObserveQuery("FraudRepo.Update", start, err)
	return err
}
