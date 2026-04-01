package dto

import "time"

type SendTransferRequest struct {
	FromAccountID string `json:"from_account_id"`
	ToAccountID   string `json:"to_account_id"`
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	Description   string `json:"description"`
}

type TransferEventResponse struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Version   int         `json:"version"`
	OccuredAt time.Time   `json:"occurred_at"`
}

type TransferResponse struct {
	ID            string    `json:"id"`
	FromAccountID string    `json:"from_account_id"`
	ToAccountID   string    `json:"to_account_id"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	Description   string    `json:"description"`
	FailureReason string    `json:"failure_reason,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
