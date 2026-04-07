package domain

import "time"

// EventType - discriminator for transfer domain events
type EventType string

const (
	EventTransferCreated   EventType = "TransferCreated"
	EventTransferCompleted EventType = "TransferCompleted"
	EventTransferFailed    EventType = "TransferFailed"
)

// Event - a single domain event for the Transfer aggregate
type Event struct {
	ID          string
	AggregateID string
	Type        EventType
	Data        EventData
	Version     int
	OccurredAt  time.Time
}

// EventData - interface that all event payloads implement
type EventData interface {
	eventData() // unexported marker - only this package can implement
}

// --- Event payloads ---

type TransferCreatedData struct {
	FromAccountID string
	ToAccountID   string
	Amount        int64
	Currency      string
	Description   string
}

type TransferCompletedData struct{}

type TransferFailedData struct {
	Reason string
}

// Marker method implementations
func (d TransferCreatedData) eventData()   {}
func (d TransferCompletedData) eventData() {}
func (d TransferFailedData) eventData()    {}
