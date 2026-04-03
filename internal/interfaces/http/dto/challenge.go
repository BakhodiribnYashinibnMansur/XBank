package dto

import "time"

type ChallengeRequestDTO struct {
	Method   string `json:"method"`   // PASSWORD or TOTP
	Action   string `json:"action"`   // e.g. "transfer", "card_issue"
	Metadata string `json:"metadata"` // e.g. transfer amount/details
}

type ChallengeVerifyDTO struct {
	ChallengeID string `json:"challenge_id"`
	Password    string `json:"password"` // for PASSWORD method
}

type ChallengeResponse struct {
	ID        string    `json:"id"`
	Method    string    `json:"method"`
	Status    string    `json:"status"`
	Action    string    `json:"action"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ChallengeVerifiedResponse struct {
	ChallengeID string    `json:"challenge_id"`
	Token       string    `json:"token"`
	ExpiresAt   time.Time `json:"expires_at"`
}
