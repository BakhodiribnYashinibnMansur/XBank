package domain

import "time"

// DomainEvent represents something that happened in the domain.
// Events are immutable facts — they describe state transitions that already occurred.
type DomainEvent interface {
	// EventName returns a unique, dot-separated event name (e.g. "user.created").
	EventName() string
	// OccurredAt returns when the event happened.
	OccurredAt() time.Time
	// AggregateID returns the ID of the aggregate that produced this event.
	AggregateID() string
}

// BaseEvent provides common fields for domain events.
type BaseEvent struct {
	ID          string
	AggrID      string
	Name        string
	OccurredAtT time.Time
}

func NewBaseEvent(name, aggregateID string) BaseEvent {
	return BaseEvent{
		AggrID:      aggregateID,
		Name:        name,
		OccurredAtT: time.Now(),
	}
}

func (e BaseEvent) EventName() string    { return e.Name }
func (e BaseEvent) OccurredAt() time.Time { return e.OccurredAtT }
func (e BaseEvent) AggregateID() string  { return e.AggrID }
