package postgres

import (
	"context"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/supporting/statistics/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SnapshotRepo persists daily KPI snapshots.
type SnapshotRepo struct {
	pool *pgxpool.Pool
}

func NewSnapshotRepo(pool *pgxpool.Pool) *SnapshotRepo {
	return &SnapshotRepo{pool: pool}
}

func (r *SnapshotRepo) Save(ctx context.Context, s *domain.DailySnapshot) error {
	start := time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO statistics_snapshots
		 (id, date, total_users, total_accounts, active_accounts, total_transfers,
		  total_cards, pending_kyc, flagged_fraud, system_errors, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 ON CONFLICT (date) DO UPDATE SET
		  total_users=$3, total_accounts=$4, active_accounts=$5, total_transfers=$6,
		  total_cards=$7, pending_kyc=$8, flagged_fraud=$9, system_errors=$10`,
		s.ID, s.Date, s.TotalUsers, s.TotalAccounts, s.ActiveAccounts, s.TotalTransfers,
		s.TotalCards, s.PendingKYC, s.FlaggedFraud, s.SystemErrors, time.Now(),
	)
	metrics.ObserveQuery("StatisticsSnapshotRepo.Save", start, err)
	return err
}
