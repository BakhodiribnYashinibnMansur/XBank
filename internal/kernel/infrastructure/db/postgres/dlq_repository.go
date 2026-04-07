package postgres

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DLQEntry - a single dead letter queue message
type DLQEntry struct {
	ID           int64
	Topic        string
	PartitionKey string
	Payload      []byte
	ErrorMsg     string
	RetryCount   int
	MaxRetries   int
	Status       string // PENDING, DELIVERED, DEAD
	NextRetry    time.Time
	CreatedAt    time.Time
}

// DLQRepository - Dead Letter Queue persistence
type DLQRepository struct {
	pool *pgxpool.Pool
}

func NewDLQRepository(pool *pgxpool.Pool) *DLQRepository {
	return &DLQRepository{pool: pool}
}

// Insert - store a failed message in the DLQ
func (r *DLQRepository) Insert(ctx context.Context, topic, partitionKey string, payload []byte, errMsg string) error {
	start := time.Now()
	_, err := r.pool.Exec(ctx,
		`INSERT INTO dead_letter_queue (topic, partition_key, payload, error_msg, next_retry)
		 VALUES ($1, $2, $3, $4, NOW() + INTERVAL '1 minute')`,
		topic, partitionKey, payload, errMsg,
	)
	metrics.ObserveQuery("DLQ.Insert", start, err)
	return err
}

// FetchPending - get messages ready for retry (up to limit)
func (r *DLQRepository) FetchPending(ctx context.Context, limit int) ([]DLQEntry, error) {
	start := time.Now()
	rows, err := r.pool.Query(ctx,
		`SELECT id, topic, partition_key, payload, error_msg, retry_count, max_retries, status, next_retry, created_at
		 FROM dead_letter_queue
		 WHERE status = 'PENDING' AND next_retry <= NOW()
		 ORDER BY next_retry ASC
		 LIMIT $1
		 FOR UPDATE SKIP LOCKED`,
		limit,
	)
	if err != nil {
		metrics.ObserveQuery("DLQ.FetchPending", start, err)
		return nil, err
	}
	defer rows.Close()

	var entries []DLQEntry
	for rows.Next() {
		var e DLQEntry
		if err := rows.Scan(&e.ID, &e.Topic, &e.PartitionKey, &e.Payload, &e.ErrorMsg,
			&e.RetryCount, &e.MaxRetries, &e.Status, &e.NextRetry, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}

	metrics.ObserveQuery("DLQ.FetchPending", start, nil)
	return entries, nil
}

// MarkDelivered - message successfully published on retry
func (r *DLQRepository) MarkDelivered(ctx context.Context, id int64) error {
	start := time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE dead_letter_queue SET status = 'DELIVERED', delivered_at = NOW() WHERE id = $1`,
		id,
	)
	metrics.ObserveQuery("DLQ.MarkDelivered", start, err)
	return err
}

// IncrementRetry - bump retry count and set next retry with exponential backoff.
// If max retries exceeded, mark as DEAD.
func (r *DLQRepository) IncrementRetry(ctx context.Context, id int64, errMsg string) error {
	start := time.Now()
	_, err := r.pool.Exec(ctx,
		`UPDATE dead_letter_queue
		 SET retry_count = retry_count + 1,
		     error_msg = $2,
		     next_retry = NOW() + (INTERVAL '1 minute' * POWER(2, retry_count)),
		     status = CASE WHEN retry_count + 1 >= max_retries THEN 'DEAD' ELSE 'PENDING' END
		 WHERE id = $1`,
		id, errMsg,
	)
	metrics.ObserveQuery("DLQ.IncrementRetry", start, err)
	return err
}
