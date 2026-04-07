package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// MessageHandler processes a single Kafka message.
// Implementations are domain-specific (account projections, notifications, etc.).
type MessageHandler interface {
	// Handle processes a consumed message. Return nil to commit, error to retry/DLQ.
	Handle(ctx context.Context, topic string, key []byte, value []byte) error
}

// Consumer reads messages from Kafka topics and dispatches to handlers.
type Consumer struct {
	brokers  []string
	groupID  string
	handlers map[string]MessageHandler // topic → handler
	readers  []*kafka.Reader
}

// NewConsumer creates a Kafka consumer for the given consumer group.
func NewConsumer(brokers []string, groupID string) *Consumer {
	return &Consumer{
		brokers:  brokers,
		groupID:  groupID,
		handlers: make(map[string]MessageHandler),
	}
}

// Subscribe registers a handler for a topic.
func (c *Consumer) Subscribe(topic string, handler MessageHandler) {
	c.handlers[topic] = handler
}

// Start begins consuming all subscribed topics. Blocks until ctx is cancelled.
// Each topic gets its own reader goroutine for independent consumption.
func (c *Consumer) Start(ctx context.Context) {
	for topic, handler := range c.handlers {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:        c.brokers,
			Topic:          topic,
			GroupID:        c.groupID,
			MinBytes:       1e3,  // 1 KB
			MaxBytes:       10e6, // 10 MB
			MaxWait:        3 * time.Second,
			CommitInterval: time.Second,
			StartOffset:    kafka.LastOffset,
		})
		c.readers = append(c.readers, reader)

		go c.consume(ctx, reader, topic, handler)
	}
}

// consume is the per-topic read loop.
func (c *Consumer) consume(ctx context.Context, reader *kafka.Reader, topic string, handler MessageHandler) {
	logger.Log.Info("Kafka consumer started", zap.String("topic", topic), zap.String("group", c.groupID))

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				logger.Log.Info("Kafka consumer stopping", zap.String("topic", topic))
				return
			}
			logger.Log.Error("Kafka fetch error", zap.String("topic", topic), zap.Error(err))
			time.Sleep(time.Second) // backoff on transient errors
			continue
		}

		start := time.Now()
		if err := handler.Handle(ctx, topic, msg.Key, msg.Value); err != nil {
			logger.Log.Error("Message handler failed",
				zap.String("topic", topic),
				zap.String("key", string(msg.Key)),
				zap.Int64("offset", msg.Offset),
				zap.Error(err),
			)
			metrics.KafkaMessagesTotal.WithLabelValues(topic, "consume_error").Inc()
			// Commit anyway to avoid infinite retry on poison messages.
			// The handler should route failures to DLQ internally.
		} else {
			metrics.KafkaMessagesTotal.WithLabelValues(topic, "consumed").Inc()
		}
		metrics.KafkaPublishDuration.WithLabelValues(fmt.Sprintf("%s_consume", topic)).
			Observe(time.Since(start).Seconds())

		if err := reader.CommitMessages(ctx, msg); err != nil {
			logger.Log.Error("Kafka commit error", zap.String("topic", topic), zap.Error(err))
		}
	}
}

// Close gracefully shuts down all readers.
func (c *Consumer) Close() error {
	var lastErr error
	for _, r := range c.readers {
		if err := r.Close(); err != nil {
			lastErr = err
			logger.Log.Error("Kafka reader close error", zap.Error(err))
		}
	}
	return lastErr
}
