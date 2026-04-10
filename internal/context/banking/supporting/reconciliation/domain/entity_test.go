package domain

import (
	"testing"
	"time"
)

func TestReconciliationRun_Fields(t *testing.T) {
	run := &ReconciliationRun{
		ID:           "run-001",
		UserID:       "user-123",
		TotalChecked: 100,
		Matches:      98,
		Mismatches:   2,
		Status:       "COMPLETED",
		CreatedAt:    time.Now(),
	}

	if run.TotalChecked != run.Matches+run.Mismatches {
		t.Errorf("total (%d) != matches (%d) + mismatches (%d)", run.TotalChecked, run.Matches, run.Mismatches)
	}
	if run.Status != "COMPLETED" {
		t.Errorf("status = %q, want COMPLETED", run.Status)
	}
}

func TestReconciliationResult_Match(t *testing.T) {
	tests := []struct {
		name    string
		proj    int64
		ledger  int64
		isMatch bool
	}{
		{"exact match", 1000000, 1000000, true},
		{"mismatch", 1000000, 999999, false},
		{"both zero", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReconciliationResult{
				AccountID:        "acc-1",
				ProjectedBalance: tt.proj,
				LedgerBalance:    tt.ledger,
				Match:            tt.proj == tt.ledger,
			}
			if result.Match != tt.isMatch {
				t.Errorf("match = %v, want %v", result.Match, tt.isMatch)
			}
		})
	}
}
