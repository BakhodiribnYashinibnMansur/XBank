package http

import "time"

type KYCSubmitRequest struct {
	DocumentType   string `json:"document_type"`
	DocumentNumber string `json:"document_number"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	DateOfBirth    string `json:"date_of_birth"`
}

type KYCReviewRequest struct {
	VerificationID string `json:"verification_id"`
	Reason         string `json:"reason"`
}

type KYCResponse struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	DocumentType   string    `json:"document_type"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	Status         string    `json:"status"`
	RejectedReason string    `json:"rejected_reason,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}
