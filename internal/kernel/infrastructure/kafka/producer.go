package kafka

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/circuitbreaker"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/segmentio/kafka-go"
)

// Producer - Kafka event publisher implementing domain.EventPublisher
type Producer struct {
	writer  *kafka.Writer
	breaker *circuitbreaker.Breaker
}

// NewProducer creates a Kafka producer for the given brokers.
// Topic is set per-message, not per-writer.
// Circuit breaker opens after 5 consecutive failures, resets after 30s.
func NewProducer(brokers []string) *Producer {
	w := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Balancer:     &kafka.Hash{}, // partition by key
		RequiredAcks: kafka.RequireOne,
		Async:        false, // synchronous for consistency
	}
	return &Producer{
		writer:  w,
		breaker: circuitbreaker.New("kafka-producer", 5, 30*time.Second),
	}
}

// Publish sends a message to the specified Kafka topic.
// Protected by a circuit breaker to prevent cascading failures
// when Kafka is unreachable.
func (p *Producer) Publish(ctx context.Context, topic string, key string, payload []byte) error {
	start := time.Now()

	err := p.breaker.Execute(func() error {
		return p.writer.WriteMessages(ctx, kafka.Message{
			Topic: topic,
			Key:   []byte(key),
			Value: payload,
		})
	})

	status := "ok"
	if err != nil {
		status = "error"
	}
	metrics.KafkaMessagesTotal.WithLabelValues(topic, status).Inc()
	metrics.KafkaPublishDuration.WithLabelValues(topic).Observe(time.Since(start).Seconds())

	return err
}

// Close gracefully shuts down the Kafka writer.
func (p *Producer) Close() error {
	return p.writer.Close()
}
