package events

import "time"

// FraudFlagged is published when a transaction is flagged as suspicious.
type FraudFlagged struct {
	CheckID    string    `json:"check_id"`
	UserID     string    `json:"user_id"`
	Reason     string    `json:"reason"`
	OccurredAt time.Time `json:"occurred_at"`
}
