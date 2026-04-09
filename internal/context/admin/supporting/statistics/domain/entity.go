package domain

import "time"

// DailySnapshot stores a point-in-time snapshot of system KPIs.
// Used for trend analysis and historical reporting.
type DailySnapshot struct {
	ID             string
	Date           time.Time
	TotalUsers     int64
	TotalAccounts  int64
	ActiveAccounts int64
	TotalTransfers int64
	TotalCards     int64
	PendingKYC     int64
	FlaggedFraud   int64
	SystemErrors   int64
	CreatedAt      time.Time
}

// Repository persists and retrieves daily snapshots.
type Repository interface {
	// Save stores a daily snapshot (upsert by date).
	Save(date time.Time, snapshot *DailySnapshot) error
}
