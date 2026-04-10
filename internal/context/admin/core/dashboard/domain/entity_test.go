package domain

import (
	"testing"
	"time"
)

func TestOverviewStats_Fields(t *testing.T) {
	stats := &OverviewStats{
		TotalUsers:       1000,
		ActiveUsers:      500,
		TotalAccounts:    1200,
		TotalTransfers:   350,
		TotalDeposits:    200,
		TotalWithdrawals: 150,
		GeneratedAt:      time.Now(),
	}

	if stats.TotalUsers != 1000 {
		t.Errorf("TotalUsers = %d, want 1000", stats.TotalUsers)
	}
	if stats.ActiveUsers != 500 {
		t.Errorf("ActiveUsers = %d, want 500", stats.ActiveUsers)
	}
	if stats.TotalAccounts != 1200 {
		t.Errorf("TotalAccounts = %d, want 1200", stats.TotalAccounts)
	}
	if stats.GeneratedAt.IsZero() {
		t.Error("GeneratedAt should not be zero")
	}
}

func TestPeriodStats_Fields(t *testing.T) {
	now := time.Now()
	stats := &PeriodStats{
		Period:       "daily",
		StartDate:    now.Add(-24 * time.Hour),
		EndDate:      now,
		NewUsers:     50,
		NewAccounts:  30,
		Transactions: 120,
		Volume:       5_000_000,
	}

	if stats.Period != "daily" {
		t.Errorf("Period = %q, want %q", stats.Period, "daily")
	}
	if stats.NewUsers != 50 {
		t.Errorf("NewUsers = %d, want 50", stats.NewUsers)
	}
	if stats.Volume != 5_000_000 {
		t.Errorf("Volume = %d, want 5000000", stats.Volume)
	}
}
