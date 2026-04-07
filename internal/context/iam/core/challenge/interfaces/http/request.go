package http

import "time"

type ChallengeRequestDTO struct {
	Method   string `json:"method"`
	Action   string `json:"action"`
	Metadata string `json:"metadata"`
}

type ChallengeVerifyDTO struct {
	ChallengeID string `json:"challenge_id"`
	Password    string `json:"password"`
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
