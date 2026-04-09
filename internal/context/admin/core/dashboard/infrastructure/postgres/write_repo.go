package postgres

import (
	"context"
	"fmt"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/core/dashboard/domain"
	sharedpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

// WriteRepo implements dashboard.Repository using PostgreSQL.
type WriteRepo struct {
	pool *pgxpool.Pool
}

// NewWriteRepo creates a new dashboard postgres repository.
func NewWriteRepo(pool *pgxpool.Pool) *WriteRepo {
	return &WriteRepo{pool: pool}
}

// GetOverview returns aggregated system-wide statistics from multiple tables.
func (r *WriteRepo) GetOverview(ctx context.Context) (*domain.OverviewStats, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)

	stats := &domain.OverviewStats{GeneratedAt: time.Now()}

	err := db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)
	if err != nil {
		metrics.ObserveQuery("DashboardRepo.GetOverview.Users", start, err)
		return nil, fmt.Errorf("dashboard_repo: count users: %w", err)
	}

	err = db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE status = 'ACTIVE'`).Scan(&stats.ActiveUsers)
	if err != nil {
		metrics.ObserveQuery("DashboardRepo.GetOverview.ActiveUsers", start, err)
		return nil, fmt.Errorf("dashboard_repo: count active users: %w", err)
	}

	err = db.QueryRow(ctx, `SELECT COUNT(*) FROM accounts`).Scan(&stats.TotalAccounts)
	if err != nil {
		metrics.ObserveQuery("DashboardRepo.GetOverview.Accounts", start, err)
		return nil, fmt.Errorf("dashboard_repo: count accounts: %w", err)
	}

	err = db.QueryRow(ctx, `SELECT COUNT(*) FROM transfers`).Scan(&stats.TotalTransfers)
	if err != nil {
		metrics.ObserveQuery("DashboardRepo.GetOverview.Transfers", start, err)
		return nil, fmt.Errorf("dashboard_repo: count transfers: %w", err)
	}

	err = db.QueryRow(ctx,
		`SELECT COALESCE(SUM(CASE WHEN type = 'DEPOSIT' THEN amount ELSE 0 END), 0),
		        COALESCE(SUM(CASE WHEN type = 'WITHDRAWAL' THEN amount ELSE 0 END), 0)
		 FROM transfers`,
	).Scan(&stats.TotalDeposits, &stats.TotalWithdrawals)
	if err != nil {
		metrics.ObserveQuery("DashboardRepo.GetOverview.Volumes", start, err)
		return nil, fmt.Errorf("dashboard_repo: sum volumes: %w", err)
	}

	metrics.ObserveQuery("DashboardRepo.GetOverview", start, nil)
	return stats, nil
}

// GetPeriodStats returns statistics for a specific time period.
func (r *WriteRepo) GetPeriodStats(ctx context.Context, period string, startDate, endDate time.Time) (*domain.PeriodStats, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)

	stats := &domain.PeriodStats{
		Period:    period,
		StartDate: startDate,
		EndDate:   endDate,
	}

	err := db.QueryRow(ctx,
		`SELECT COUNT(*) FROM users WHERE created_at >= $1 AND created_at < $2`,
		startDate, endDate,
	).Scan(&stats.NewUsers)
	if err != nil {
		metrics.ObserveQuery("DashboardRepo.GetPeriodStats.NewUsers", start, err)
		return nil, fmt.Errorf("dashboard_repo: period new users: %w", err)
	}

	err = db.QueryRow(ctx,
		`SELECT COUNT(*) FROM accounts WHERE created_at >= $1 AND created_at < $2`,
		startDate, endDate,
	).Scan(&stats.NewAccounts)
	if err != nil {
		metrics.ObserveQuery("DashboardRepo.GetPeriodStats.NewAccounts", start, err)
		return nil, fmt.Errorf("dashboard_repo: period new accounts: %w", err)
	}

	err = db.QueryRow(ctx,
		`SELECT COUNT(*), COALESCE(SUM(amount), 0)
		 FROM transfers WHERE created_at >= $1 AND created_at < $2`,
		startDate, endDate,
	).Scan(&stats.Transactions, &stats.Volume)
	if err != nil {
		metrics.ObserveQuery("DashboardRepo.GetPeriodStats.Transfers", start, err)
		return nil, fmt.Errorf("dashboard_repo: period transfers: %w", err)
	}

	metrics.ObserveQuery("DashboardRepo.GetPeriodStats", start, nil)
	return stats, nil
}
