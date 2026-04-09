package postgres

import (
	"context"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/reconciliation/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RunRepo persists reconciliation run history.
type RunRepo struct {
	pool *pgxpool.Pool
}

func NewRunRepo(pool *pgxpool.Pool) *RunRepo {
	return &RunRepo{pool: pool}
}

// Save persists a reconciliation run record.
func (r *RunRepo) Save(ctx context.Context, run *domain.ReconciliationRun) error {
	start := time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO reconciliation_runs (id, user_id, total_checked, matches, mismatches, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		run.ID, run.UserID, run.TotalChecked, run.Matches, run.Mismatches, run.Status, run.CreatedAt,
	)
	metrics.ObserveQuery("ReconciliationRunRepo.Save", start, err)
	return err
}
