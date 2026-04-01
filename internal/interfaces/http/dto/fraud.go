package dto

import "time"

type FraudCheckResponse struct {
	ID            string    `json:"id"`
	TransferID    string    `json:"transfer_id"`
	UserID        string    `json:"user_id"`
	RiskScore     int       `json:"risk_score"`
	RiskLevel     string    `json:"risk_level"`
	Action        string    `json:"action"`
	Flags         []string  `json:"flags"`
	ReviewedBy    string    `json:"reviewed_by,omitempty"`
	ReviewComment string    `json:"review_comment,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
