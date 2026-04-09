package events

import "time"

// KYCSubmitted is published when a KYC verification is submitted.
type KYCSubmitted struct {
	VerificationID string    `json:"verification_id"`
	UserID         string    `json:"user_id"`
	DocumentType   string    `json:"document_type"`
	OccurredAt     time.Time `json:"occurred_at"`
}

// KYCApproved is published when a user's KYC verification is approved.
type KYCApproved struct {
	VerificationID string    `json:"verification_id"`
	UserID         string    `json:"user_id"`
	Level          string    `json:"level"`
	OccurredAt     time.Time `json:"occurred_at"`
}

// KYCRejected is published when a user's KYC verification is rejected.
type KYCRejected struct {
	VerificationID string    `json:"verification_id"`
	UserID         string    `json:"user_id"`
	Reason         string    `json:"reason"`
	OccurredAt     time.Time `json:"occurred_at"`
}
