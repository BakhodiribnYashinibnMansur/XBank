package mock

import (
	"context"
	"sync"
)

// EventPublisher - in-memory mock for domain.EventPublisher
type EventPublisher struct {
	mu       sync.RWMutex
	Messages []PublishedMessage
}

type PublishedMessage struct {
	Topic   string
	Key     string
	Payload []byte
}

func NewEventPublisher() *EventPublisher {
	return &EventPublisher{}
}

func (p *EventPublisher) Publish(_ context.Context, topic string, key string, payload []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Messages = append(p.Messages, PublishedMessage{Topic: topic, Key: key, Payload: payload})
	return nil
}
