package query

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// OverviewStats aggregates high-level KPIs for admin dashboard.
type OverviewStats struct {
	TotalUsers      int64 `json:"total_users"`
	TotalAccounts   int64 `json:"total_accounts"`
	ActiveAccounts  int64 `json:"active_accounts"`
	TotalTransfers  int64 `json:"total_transfers"`
	TotalCards      int64 `json:"total_cards"`
	PendingKYC      int64 `json:"pending_kyc"`
	FlaggedFraud    int64 `json:"flagged_fraud"`
	SystemErrors    int64 `json:"system_errors_pending"`
}

type OverviewHandler struct{ pool *pgxpool.Pool }

func NewOverviewHandler(pool *pgxpool.Pool) *OverviewHandler {
	return &OverviewHandler{pool: pool}
}

func (h *OverviewHandler) Handle(ctx context.Context) (*OverviewStats, error) {
	s := &OverviewStats{}

	queries := []struct {
		query string
		dest  *int64
	}{
		{`SELECT COUNT(*) FROM users`, &s.TotalUsers},
		{`SELECT COUNT(*) FROM accounts`, &s.TotalAccounts},
		{`SELECT COUNT(*) FROM accounts WHERE status = 'ACTIVE'`, &s.ActiveAccounts},
		{`SELECT COUNT(*) FROM transfers`, &s.TotalTransfers},
		{`SELECT COUNT(*) FROM cards`, &s.TotalCards},
		{`SELECT COUNT(*) FROM kyc_verifications WHERE status = 'PENDING'`, &s.PendingKYC},
		{`SELECT COUNT(*) FROM fraud_checks WHERE action = 'FLAG'`, &s.FlaggedFraud},
	}

	for _, q := range queries {
		// Non-critical: if table doesn't exist, default to 0
		if err := h.pool.QueryRow(ctx, q.query).Scan(q.dest); err != nil {
			*q.dest = 0
		}
	}

	// System errors from new BC table (may not exist yet)
	if err := h.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM system_errors WHERE resolution = 'PENDING'`,
	).Scan(&s.SystemErrors); err != nil {
		s.SystemErrors = 0
	}

	return s, nil
}
