package account

import "time"

// EventType - discriminator for account domain events
type EventType string

const (
	EventAccountOpened EventType = "AccountOpened"
	EventCredited      EventType = "Credited"
	EventDebited       EventType = "Debited"
	EventFrozen        EventType = "Frozen"
	EventUnfrozen      EventType = "Unfrozen"
	EventClosed        EventType = "Closed"
)

// Event - a single domain event for the Account aggregate
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

type AccountOpenedData struct {
	UserID        string
	AccountNumber string
	Currency      string
}

type CreditedData struct {
	Amount   int64
	Currency string
}

type DebitedData struct {
	Amount   int64
	Currency string
}

type FrozenData struct{}
type UnfrozenData struct{}
type ClosedData struct{}

// Marker method implementations
func (d AccountOpenedData) eventData() {}
func (d CreditedData) eventData()      {}
func (d DebitedData) eventData()       {}
func (d FrozenData) eventData()        {}
func (d UnfrozenData) eventData()      {}
func (d ClosedData) eventData()        {}
