package http

import "time"

type AddBeneficiaryRequest struct {
	Name          string `json:"name"`
	AccountNumber string `json:"account_number"`
	BankName      string `json:"bank_name"`
	BankCode      string `json:"bank_code"`
	Currency      string `json:"currency"`
	Type          string `json:"type"`
}

type BeneficiaryResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	AccountNumber string    `json:"account_number"`
	BankName      string    `json:"bank_name"`
	BankCode      string    `json:"bank_code,omitempty"`
	Currency      string    `json:"currency"`
	Type          string    `json:"type"`
	Verified      bool      `json:"verified"`
	CreatedAt     time.Time `json:"created_at"`
}
