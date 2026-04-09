package outbox

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.uber.org/zap"
)

// Relay reads pending outbox entries and publishes them to Kafka.
// After successful publish, the entry is deleted from the outbox table.
// Failed entries remain for the next relay cycle.
type Relay struct {
	repo      *Repository
	publisher domain.EventPublisher // raw Kafka producer (not outbox publisher)
}

// NewRelay creates an outbox relay worker.
// publisher must be the real Kafka producer (with DLQ fallback), not the outbox publisher.
func NewRelay(repo *Repository, publisher domain.EventPublisher) *Relay {
	return &Relay{
		repo:      repo,
		publisher: publisher,
	}
}

// ProcessBatch fetches pending outbox entries and publishes them to Kafka.
// Returns count of delivered and failed messages.
func (r *Relay) ProcessBatch(ctx context.Context, batchSize int) (delivered, failed int) {
	entries, err := r.repo.FetchBatch(ctx, batchSize)
	if err != nil {
		logger.Log.Error("outbox fetch failed", zap.Error(err))
		return 0, 0
	}

	if len(entries) == 0 {
		return 0, 0
	}

	var deliveredIDs []int64

	for _, entry := range entries {
		start := time.Now()
		err := r.publisher.Publish(ctx, entry.Topic, entry.PartitionKey, entry.Payload)
		metrics.KafkaPublishDuration.WithLabelValues(entry.Topic).Observe(time.Since(start).Seconds())

		if err == nil {
			deliveredIDs = append(deliveredIDs, entry.ID)
			delivered++
			metrics.KafkaMessagesTotal.WithLabelValues(entry.Topic, "outbox_ok").Inc()
		} else {
			failed++
			metrics.KafkaMessagesTotal.WithLabelValues(entry.Topic, "outbox_fail").Inc()
			logger.Log.Warn("outbox relay publish failed",
				zap.String("topic", entry.Topic),
				zap.String("aggregate_id", entry.AggregateID),
				zap.Error(err),
			)
		}
	}

	// Batch delete all successfully delivered entries
	if len(deliveredIDs) > 0 {
		if err := r.repo.DeleteBatch(ctx, deliveredIDs); err != nil {
			logger.Log.Error("outbox batch delete failed", zap.Error(err))
		}
	}

	logger.Log.Info("outbox relay completed",
		zap.Int("delivered", delivered),
		zap.Int("failed", failed),
	)

	return delivered, failed
}
