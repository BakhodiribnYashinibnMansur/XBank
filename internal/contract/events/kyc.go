package events

import "time"

// KYCApproved is published when a user's KYC verification is approved.
type KYCApproved struct {
	UserID     string    `json:"user_id"`
	Level      string    `json:"level"`
	OccurredAt time.Time `json:"occurred_at"`
}

// KYCRejected is published when a user's KYC verification is rejected.
type KYCRejected struct {
	UserID     string    `json:"user_id"`
	Reason     string    `json:"reason"`
	OccurredAt time.Time `json:"occurred_at"`
}
