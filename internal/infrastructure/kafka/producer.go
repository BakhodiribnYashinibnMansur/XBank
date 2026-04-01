package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
)

// Producer - Kafka event publisher implementing shared.EventPublisher
type Producer struct {
	writer *kafka.Writer
}

// NewProducer creates a Kafka producer for the given brokers.
// Topic is set per-message, not per-writer.
func NewProducer(brokers []string) *Producer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.Hash{}, // partition by key
		RequiredAcks: kafka.RequireOne,
		Async:        false, // synchronous for consistency
	}
	return &Producer{writer: w}
}

// Publish sends a message to the specified Kafka topic.
func (p *Producer) Publish(ctx context.Context, topic string, key string, payload []byte) error {
	return p.writer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: payload,
	})
}

// Close gracefully shuts down the Kafka writer.
func (p *Producer) Close() error {
	return p.writer.Close()
}
