package http

import "time"

// OverviewResponse represents the admin dashboard overview.
type OverviewResponse struct {
	TotalUsers       int64     `json:"total_users"`
	ActiveUsers      int64     `json:"active_users"`
	TotalAccounts    int64     `json:"total_accounts"`
	TotalTransfers   int64     `json:"total_transfers"`
	TotalDeposits    int64     `json:"total_deposits"`
	TotalWithdrawals int64     `json:"total_withdrawals"`
	GeneratedAt      time.Time `json:"generated_at"`
}

// PeriodStatsResponse represents statistics for a time period.
type PeriodStatsResponse struct {
	Period       string    `json:"period"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	NewUsers     int64     `json:"new_users"`
	NewAccounts  int64     `json:"new_accounts"`
	Transactions int64     `json:"transactions"`
	Volume       int64     `json:"volume"`
}
