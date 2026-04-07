package eventbus

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

// InMemoryBus is a synchronous in-memory event bus.
type InMemoryBus struct {
	handlers map[string][]func(ctx context.Context, event domain.DomainEvent) error
}

// New creates a new in-memory event bus.
func New() *InMemoryBus {
	return &InMemoryBus{
		handlers: make(map[string][]func(ctx context.Context, event domain.DomainEvent) error),
	}
}

// Publish sends events to all registered handlers.
func (b *InMemoryBus) Publish(ctx context.Context, events ...domain.DomainEvent) error {
	for _, event := range events {
		for _, h := range b.handlers[event.EventName()] {
			if err := h(ctx, event); err != nil {
				return err
			}
		}
	}
	return nil
}

// Subscribe registers a handler for the given event name.
func (b *InMemoryBus) Subscribe(eventName string, handler func(ctx context.Context, event domain.DomainEvent) error) {
	b.handlers[eventName] = append(b.handlers[eventName], handler)
}
