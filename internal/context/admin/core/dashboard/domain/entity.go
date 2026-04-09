package domain

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

var (
	ErrStatsUnavailable = domain.NewDomainError("STATS_UNAVAILABLE", "dashboard statistics are unavailable")
)

// OverviewStats holds aggregated dashboard metrics for the admin panel.
type OverviewStats struct {
	TotalUsers       int64     `json:"total_users"`
	ActiveUsers      int64     `json:"active_users"`
	TotalAccounts    int64     `json:"total_accounts"`
	TotalTransfers   int64     `json:"total_transfers"`
	TotalDeposits    int64     `json:"total_deposits"`
	TotalWithdrawals int64     `json:"total_withdrawals"`
	GeneratedAt      time.Time `json:"generated_at"`
}

// PeriodStats holds metrics for a specific time period.
type PeriodStats struct {
	Period       string    `json:"period"` // "daily", "weekly", "monthly"
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	NewUsers     int64     `json:"new_users"`
	NewAccounts  int64     `json:"new_accounts"`
	Transactions int64     `json:"transactions"`
	Volume       int64     `json:"volume"` // total transaction volume in minor units
}

// Repository defines the read interface for dashboard statistics.
type Repository interface {
	GetOverview(ctx context.Context) (*OverviewStats, error)
	GetPeriodStats(ctx context.Context, period string, start, end time.Time) (*PeriodStats, error)
}
