package dto

import "time"

type CreateAccountRequest struct {
	Currency string `json:"currency"`
}

type DepositRequest struct {
	AccountID string `json:"account_id"`
	Amount    int64  `json:"amount"`
}

type WithdrawRequest struct {
	AccountID string `json:"account_id"`
	Amount    int64  `json:"amount"`
}

type CloseAccountRequest struct {
	AccountID string `json:"account_id"`
}

type AccountResponse struct {
	ID            string    `json:"id"`
	AccountNumber string    `json:"account_number"`
	Balance       int64     `json:"balance"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}
