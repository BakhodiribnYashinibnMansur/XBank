package domain

import "time"

// ReconciliationRun records a single reconciliation execution.
type ReconciliationRun struct {
	ID         string
	UserID     string
	TotalChecked int
	Matches      int
	Mismatches   int
	Status       string // "COMPLETED", "PARTIAL_FAILURE"
	CreatedAt    time.Time
}

// ReconciliationResult is the detailed result for one account check.
type ReconciliationResult struct {
	AccountID        string `json:"account_id"`
	ProjectedBalance int64  `json:"projected_balance"`
	LedgerBalance    int64  `json:"ledger_balance"`
	Match            bool   `json:"match"`
}
