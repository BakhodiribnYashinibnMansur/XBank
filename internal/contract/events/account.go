package events

import "time"

// AccountOpened is published when a new account is created.
type AccountOpened struct {
	AccountID     string    `json:"account_id"`
	UserID        string    `json:"user_id"`
	AccountNumber string    `json:"account_number"`
	Currency      string    `json:"currency"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// AccountCredited is published when funds are deposited into an account.
type AccountCredited struct {
	AccountID  string    `json:"account_id"`
	Amount     int64     `json:"amount"`
	Currency   string    `json:"currency"`
	OccurredAt time.Time `json:"occurred_at"`
}

// AccountDebited is published when funds are withdrawn from an account.
type AccountDebited struct {
	AccountID  string    `json:"account_id"`
	Amount     int64     `json:"amount"`
	Currency   string    `json:"currency"`
	OccurredAt time.Time `json:"occurred_at"`
}

// AccountFrozen is published when an account is frozen.
type AccountFrozen struct {
	AccountID  string    `json:"account_id"`
	OccurredAt time.Time `json:"occurred_at"`
}

// AccountClosed is published when an account is closed.
type AccountClosed struct {
	AccountID  string    `json:"account_id"`
	OccurredAt time.Time `json:"occurred_at"`
}
