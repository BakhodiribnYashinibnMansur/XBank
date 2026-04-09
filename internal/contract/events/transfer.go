package events

import "time"

// TransferCreated is published when a new transfer is initiated.
type TransferCreated struct {
	TransferID    string    `json:"transfer_id"`
	FromAccountID string    `json:"from_account_id"`
	ToAccountID   string    `json:"to_account_id"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	Description   string    `json:"description,omitempty"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// TransferCompleted is published when a transfer is successfully processed.
type TransferCompleted struct {
	TransferID    string    `json:"transfer_id"`
	FromAccountID string    `json:"from_account_id"`
	ToAccountID   string    `json:"to_account_id"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	OccurredAt    time.Time `json:"occurred_at"`
}

// TransferFailed is published when a transfer fails.
type TransferFailed struct {
	TransferID    string    `json:"transfer_id"`
	FromAccountID string    `json:"from_account_id"`
	ToAccountID   string    `json:"to_account_id"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	Reason        string    `json:"reason"`
	OccurredAt    time.Time `json:"occurred_at"`
}
