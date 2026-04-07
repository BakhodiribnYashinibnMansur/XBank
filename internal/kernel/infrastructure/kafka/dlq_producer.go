package kafka

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.uber.org/zap"
)

// DLQProducer wraps a Kafka producer with dead-letter-queue fallback.
// If Kafka publish fails, the message is stored in PostgreSQL DLQ
// for later retry instead of being lost.
type DLQProducer struct {
	producer *Producer
	dlqRepo  *postgres.DLQRepository
}

// NewDLQProducer creates a Kafka producer with DLQ support.
func NewDLQProducer(producer *Producer, dlqRepo *postgres.DLQRepository) *DLQProducer {
	return &DLQProducer{
		producer: producer,
		dlqRepo:  dlqRepo,
	}
}

// Publish attempts to send to Kafka. On failure, stores in DLQ.
func (p *DLQProducer) Publish(ctx context.Context, topic string, key string, payload []byte) error {
	err := p.producer.Publish(ctx, topic, key, payload)
	if err == nil {
		return nil
	}

	// Kafka failed — store in DLQ for retry
	logger.Log.Warn("Kafka publish failed, storing in DLQ",
		zap.String("topic", topic),
		zap.String("key", key),
		zap.Error(err),
	)

	if dlqErr := p.dlqRepo.Insert(ctx, topic, key, payload, err.Error()); dlqErr != nil {
		logger.Log.Error("DLQ insert also failed — message lost",
			zap.String("topic", topic),
			zap.Error(dlqErr),
		)
		return dlqErr
	}

	metrics.KafkaMessagesTotal.WithLabelValues(topic, "dlq").Inc()
	return nil // DLQ saved successfully, caller should not fail
}

// Close gracefully shuts down the underlying Kafka producer.
func (p *DLQProducer) Close() error {
	return p.producer.Close()
}

// RetryPending fetches pending DLQ messages and attempts republish.
// Call this from a background worker or pg_cron triggered endpoint.
func (p *DLQProducer) RetryPending(ctx context.Context, batchSize int) (delivered, failed int) {
	entries, err := p.dlqRepo.FetchPending(ctx, batchSize)
	if err != nil {
		logger.Log.Error("DLQ fetch failed", zap.Error(err))
		return 0, 0
	}

	for _, entry := range entries {
		start := time.Now()
		err := p.producer.Publish(ctx, entry.Topic, entry.PartitionKey, entry.Payload)
		if err == nil {
			p.dlqRepo.MarkDelivered(ctx, entry.ID)
			delivered++
			metrics.KafkaMessagesTotal.WithLabelValues(entry.Topic, "dlq_retry_ok").Inc()
		} else {
			p.dlqRepo.IncrementRetry(ctx, entry.ID, err.Error())
			failed++
			metrics.KafkaMessagesTotal.WithLabelValues(entry.Topic, "dlq_retry_fail").Inc()
		}
		metrics.KafkaPublishDuration.WithLabelValues(entry.Topic).Observe(time.Since(start).Seconds())
	}

	if delivered+failed > 0 {
		logger.Log.Info("DLQ retry completed",
			zap.Int("delivered", delivered),
			zap.Int("failed", failed),
		)
	}

	return delivered, failed
}
