package events

import "time"

// CardIssued is published when a new card is created.
type CardIssued struct {
	CardID    string    `json:"card_id"`
	AccountID string    `json:"account_id"`
	UserID    string    `json:"user_id"`
	CardType  string    `json:"card_type"`
	OccurredAt time.Time `json:"occurred_at"`
}

// CardBlocked is published when a card is blocked.
type CardBlocked struct {
	CardID     string    `json:"card_id"`
	Reason     string    `json:"reason"`
	OccurredAt time.Time `json:"occurred_at"`
}

// CardActivated is published when a card is activated.
type CardActivated struct {
	CardID     string    `json:"card_id"`
	OccurredAt time.Time `json:"occurred_at"`
}
