package dto

import "time"

type IssueCardRequest struct {
	AccountID string `json:"account_id"`
	CardType  string `json:"card_type"` // DEBIT, VIRTUAL
}

type ActivateCardRequest struct {
	PIN string `json:"pin"` // 4 digits
}

type VerifyPINRequest struct {
	PIN string `json:"pin"`
}

type ChangePINRequest struct {
	OldPIN string `json:"old_pin"`
	NewPIN string `json:"new_pin"`
}

type CardResponse struct {
	ID          string    `json:"id"`
	AccountID   string    `json:"account_id"`
	MaskedPAN   string    `json:"masked_pan"` // PAN never returned in full!
	ExpiryMonth int       `json:"expiry_month"`
	ExpiryYear  int       `json:"expiry_year"`
	CardType    string    `json:"card_type"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}
