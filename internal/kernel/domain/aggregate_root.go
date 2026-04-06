package domain

import "time"

// AggregateRoot is the base for all aggregate roots in the system.
// It collects domain events that occurred during command execution
// and provides common entity fields.
type AggregateRoot struct {
	BaseEntity
	events []DomainEvent
}

// AddEvent records a domain event to be published after persistence.
func (a *AggregateRoot) AddEvent(event DomainEvent) {
	a.events = append(a.events, event)
}

// PullEvents returns all uncommitted domain events and clears the internal list.
// Call this after successful persistence to publish events via EventBus.
func (a *AggregateRoot) PullEvents() []DomainEvent {
	events := a.events
	a.events = nil
	return events
}

// HasEvents returns true if there are uncommitted domain events.
func (a *AggregateRoot) HasEvents() bool {
	return len(a.events) > 0
}

// BaseEntity provides common fields for all entities.
type BaseEntity struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time // nil = not deleted (soft delete)
}

// IsDeleted returns true if the entity has been soft-deleted.
func (e *BaseEntity) IsDeleted() bool {
	return e.DeletedAt != nil
}

// MarkDeleted performs a soft delete by setting DeletedAt.
func (e *BaseEntity) MarkDeleted() {
	now := time.Now()
	e.DeletedAt = &now
	e.UpdatedAt = now
}

// Touch updates the UpdatedAt timestamp.
func (e *BaseEntity) Touch() {
	e.UpdatedAt = time.Now()
}
