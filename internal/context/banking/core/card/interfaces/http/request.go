package http

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

type Enroll3DSRequest struct {
	Version string `json:"version"` // "2.1", "2.2"
}

type SetEMVRequest struct {
	AID string `json:"aid"` // e.g. "A0000000041010"
}

type HoldRequest struct {
	CardID    string `json:"card_id"`
	AccountID string `json:"account_id"`
	Merchant  string `json:"merchant"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
}

type HoldResponse struct {
	ID        string `json:"id"`
	CardID    string `json:"card_id"`
	AccountID string `json:"account_id"`
	Merchant  string `json:"merchant"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
	Status    string `json:"status"`
	HeldAt    string `json:"held_at"`
	ExpiresAt string `json:"expires_at"`
}

type TokenResponse struct {
	Token    string `json:"token"`
	CardID   string `json:"card_id"`
	LastFour string `json:"last_four"`
	IsActive bool   `json:"is_active"`
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
