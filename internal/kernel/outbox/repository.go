package outbox

import (
	"context"
	"time"

	sharedpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Entry represents a single outbox message.
type Entry struct {
	ID            int64
	AggregateType string
	AggregateID   string
	Topic         string
	PartitionKey  string
	Payload       []byte
	CreatedAt     time.Time
}

// Repository provides persistence for the transactional outbox.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates an outbox repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Insert stores a message in the outbox within the current transaction.
func (r *Repository) Insert(ctx context.Context, entry Entry) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)

	_, err := db.Exec(ctx,
		`INSERT INTO outbox (aggregate_type, aggregate_id, topic, partition_key, payload)
		 VALUES ($1, $2, $3, $4, $5)`,
		entry.AggregateType, entry.AggregateID,
		entry.Topic, entry.PartitionKey, entry.Payload,
	)
	metrics.ObserveQuery("Outbox.Insert", start, err)
	return err
}

// FetchBatch retrieves pending outbox entries for relay processing.
// Uses FOR UPDATE SKIP LOCKED to allow concurrent relay workers.
func (r *Repository) FetchBatch(ctx context.Context, limit int) ([]Entry, error) {
	start := time.Now()

	rows, err := r.pool.Query(ctx,
		`SELECT id, aggregate_type, aggregate_id, topic, partition_key, payload, created_at
		 FROM outbox
		 ORDER BY id ASC
		 LIMIT $1
		 FOR UPDATE SKIP LOCKED`,
		limit,
	)
	if err != nil {
		metrics.ObserveQuery("Outbox.FetchBatch", start, err)
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.AggregateType, &e.AggregateID,
			&e.Topic, &e.PartitionKey, &e.Payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}

	metrics.ObserveQuery("Outbox.FetchBatch", start, nil)
	return entries, nil
}

// Delete removes a successfully published outbox entry.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	start := time.Now()
	_, err := r.pool.Exec(ctx, `DELETE FROM outbox WHERE id = $1`, id)
	metrics.ObserveQuery("Outbox.Delete", start, err)
	return err
}

// DeleteBatch removes multiple outbox entries by IDs.
func (r *Repository) DeleteBatch(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	start := time.Now()
	_, err := r.pool.Exec(ctx, `DELETE FROM outbox WHERE id = ANY($1)`, ids)
	metrics.ObserveQuery("Outbox.DeleteBatch", start, err)
	return err
}
