package notification

import "time"

type EventType string

const (
	EventTransferCompleted EventType = "transfer.completed"
	EventTransferReceived  EventType = "transfer.received"
	EventTransferFailed    EventType = "transfer.failed"
	EventCardBlocked       EventType = "card.blocked"
	EventCardActivated     EventType = "card.activated"
	EventLoginNew          EventType = "session.new_login"
	EventKYCApproved       EventType = "kyc.approved"
	EventKYCRejected       EventType = "kyc.rejected"
	EventAMLFlagged        EventType = "aml.flagged"
)

// Event - a notification event sent to the client via SSE
type Event struct {
	ID        string            `json:"id"`
	UserID    string            `json:"user_id"`
	Type      EventType         `json:"type"`
	Title     string            `json:"title"`
	Message   string            `json:"message"`
	Data      map[string]string `json:"data,omitempty"`
	Read      bool              `json:"read"`
	CreatedAt time.Time         `json:"created_at"`
}
