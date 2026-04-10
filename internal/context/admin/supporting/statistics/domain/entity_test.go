package domain

import (
	"testing"
	"time"
)

func TestDailySnapshot_Fields(t *testing.T) {
	snap := &DailySnapshot{
		ID:             "snap-001",
		Date:           time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC),
		TotalUsers:     5000,
		TotalAccounts:  8000,
		ActiveAccounts: 6500,
		TotalTransfers: 25000,
		TotalCards:     3000,
		PendingKYC:     150,
		FlaggedFraud:   12,
		SystemErrors:   3,
		CreatedAt:      time.Now(),
	}

	if snap.TotalUsers != 5000 {
		t.Errorf("TotalUsers = %d, want 5000", snap.TotalUsers)
	}
	if snap.ActiveAccounts > snap.TotalAccounts {
		t.Errorf("ActiveAccounts (%d) > TotalAccounts (%d)", snap.ActiveAccounts, snap.TotalAccounts)
	}
	if snap.Date.Year() != 2026 {
		t.Errorf("Date year = %d, want 2026", snap.Date.Year())
	}
	if snap.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestDailySnapshot_ZeroValues(t *testing.T) {
	snap := &DailySnapshot{
		Date: time.Now(),
	}

	if snap.TotalUsers != 0 {
		t.Errorf("default TotalUsers = %d, want 0", snap.TotalUsers)
	}
	if snap.FlaggedFraud != 0 {
		t.Errorf("default FlaggedFraud = %d, want 0", snap.FlaggedFraud)
	}
}
