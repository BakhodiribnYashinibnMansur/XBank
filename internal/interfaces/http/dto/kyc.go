package dto

import "time"

type KYCSubmitRequest struct {
	DocumentType   string `json:"document_type"` // PASSPORT, ID_CARD, DRIVER_LICENSE
	DocumentNumber string `json:"document_number"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	DateOfBirth    string `json:"date_of_birth"` // YYYY-MM-DD
}

type KYCReviewRequest struct {
	VerificationID string `json:"verification_id"`
	Reason         string `json:"reason"` // only for reject
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
