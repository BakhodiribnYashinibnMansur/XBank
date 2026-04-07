package postgres

import (
	"context"
	"strconv"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/transfer"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/jackc/pgx/v5/pgxpool"
	sharedPG "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/postgres"
)

type TransferEventRepository struct {
	pool *pgxpool.Pool
}

func NewTransferEventRepository(pool *pgxpool.Pool) *TransferEventRepository {
	return &TransferEventRepository{pool: pool}
}

// Append - persist transfer events as EAV rows
func (r *TransferEventRepository) Append(ctx context.Context, aggregateID string, expectedVersion int, events []transfer.Event) error {
	start := time.Now()
	db := sharedPG.ExtractDBTX(ctx, r.pool)

	for _, e := range events {
		attrs := transferEventDataToAttrs(e.Data)

		// Events with no data (TransferCompleted) get a marker row
		if len(attrs) == 0 {
			attrs = map[string]string{"_": "_"}
		}

		for key, val := range attrs {
			_, err := db.Exec(ctx,
				`INSERT INTO transfer_events (aggregate_id, event_type, version, attr_key, attr_value, occurred_at)
				 VALUES ($1, $2, $3, $4, $5, $6)`,
				aggregateID, string(e.Type), e.Version, key, val, e.OccurredAt,
			)
			if err != nil {
				metrics.ObserveQuery("TransferEventRepo.Append", start, err)
				return apperror.ErrConcurrencyConflict
			}
		}
	}
	metrics.ObserveQuery("TransferEventRepo.Append", start, nil)
	return nil
}

// LoadEvents - load all transfer events, reconstruct from EAV rows
func (r *TransferEventRepository) LoadEvents(ctx context.Context, aggregateID string) ([]transfer.Event, error) {
	return r.loadEvents(ctx, aggregateID, 0)
}

// LoadEventsFromVersion - load events after a given version
func (r *TransferEventRepository) LoadEventsFromVersion(ctx context.Context, aggregateID string, fromVersion int) ([]transfer.Event, error) {
	return r.loadEvents(ctx, aggregateID, fromVersion)
}

func (r *TransferEventRepository) loadEvents(ctx context.Context, aggregateID string, fromVersion int) ([]transfer.Event, error) {
	start := time.Now()
	db := sharedPG.ExtractDBTX(ctx, r.pool)

	rows, err := db.Query(ctx,
		`SELECT aggregate_id, event_type, version, attr_key, attr_value, occurred_at
		 FROM transfer_events
		 WHERE aggregate_id = $1 AND version > $2
		 ORDER BY version ASC, attr_key ASC`,
		aggregateID, fromVersion,
	)
	if err != nil {
		metrics.ObserveQuery("TransferEventRepo.LoadEvents", start, err)
		return nil, err
	}
	defer rows.Close()

	type rawRow struct {
		aggregateID string
		eventType   string
		version     int
		key         string
		value       string
		occurredAt  time.Time
	}

	var allRows []rawRow
	for rows.Next() {
		var row rawRow
		if err := rows.Scan(&row.aggregateID, &row.eventType, &row.version, &row.key, &row.value, &row.occurredAt); err != nil {
			return nil, err
		}
		allRows = append(allRows, row)
	}

	// Group by version
	grouped := make(map[int][]rawRow)
	var versions []int
	for _, row := range allRows {
		if _, exists := grouped[row.version]; !exists {
			versions = append(versions, row.version)
		}
		grouped[row.version] = append(grouped[row.version], row)
	}

	// Reconstruct events
	var events []transfer.Event
	for _, v := range versions {
		rowGroup := grouped[v]
		if len(rowGroup) == 0 {
			continue
		}

		attrs := make(map[string]string)
		for _, row := range rowGroup {
			attrs[row.key] = row.value
		}

		first := rowGroup[0]
		eventType := transfer.EventType(first.eventType)
		data := attrsToTransferEventData(eventType, attrs)

		events = append(events, transfer.Event{
			AggregateID: first.aggregateID,
			Type:        eventType,
			Data:        data,
			Version:     first.version,
			OccurredAt:  first.occurredAt,
		})
	}

	metrics.ObserveQuery("TransferEventRepo.LoadEvents", start, nil)
	return events, nil
}

// SaveSnapshot - upsert into dedicated transfer_snapshots table
func (r *TransferEventRepository) SaveSnapshot(ctx context.Context, snapshot transfer.Snapshot) error {
	start := time.Now()
	db := sharedPG.ExtractDBTX(ctx, r.pool)

	_, err := db.Exec(ctx,
		`INSERT INTO transfer_snapshots (aggregate_id, version, from_account_id, to_account_id, amount, currency, status, description, failure_reason, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 ON CONFLICT (aggregate_id)
		 DO UPDATE SET version = EXCLUDED.version,
		               from_account_id = EXCLUDED.from_account_id,
		               to_account_id = EXCLUDED.to_account_id,
		               amount = EXCLUDED.amount,
		               currency = EXCLUDED.currency,
		               status = EXCLUDED.status,
		               description = EXCLUDED.description,
		               failure_reason = EXCLUDED.failure_reason,
		               created_at = EXCLUDED.created_at`,
		snapshot.AggregateID, snapshot.Version,
		snapshot.State.FromAccountID, snapshot.State.ToAccountID,
		snapshot.State.Amount, snapshot.State.Currency,
		snapshot.State.Status, snapshot.State.Description, snapshot.State.FailureReason,
		snapshot.CreatedAt,
	)
	if err != nil {
		metrics.ObserveQuery("TransferEventRepo.SaveSnapshot", start, err)
		return err
	}

	metrics.ObserveQuery("TransferEventRepo.SaveSnapshot", start, nil)
	return nil
}

// LoadSnapshot - load latest snapshot from dedicated table
func (r *TransferEventRepository) LoadSnapshot(ctx context.Context, aggregateID string) (*transfer.Snapshot, error) {
	start := time.Now()
	db := sharedPG.ExtractDBTX(ctx, r.pool)

	var version int
	var fromAccID, toAccID, currency, status, description, failureReason string
	var amount int64
	var createdAt time.Time

	err := db.QueryRow(ctx,
		`SELECT version, from_account_id, to_account_id, amount, currency, status, description, failure_reason, created_at
		 FROM transfer_snapshots
		 WHERE aggregate_id = $1`,
		aggregateID,
	).Scan(&version, &fromAccID, &toAccID, &amount, &currency, &status, &description, &failureReason, &createdAt)

	if err != nil {
		metrics.ObserveQuery("TransferEventRepo.LoadSnapshot", start, nil)
		return nil, nil // not found = no snapshot
	}

	metrics.ObserveQuery("TransferEventRepo.LoadSnapshot", start, nil)
	return &transfer.Snapshot{
		AggregateID: aggregateID,
		Version:     version,
		State: transfer.SnapshotState{
			FromAccountID: fromAccID,
			ToAccountID:   toAccID,
			Amount:        amount,
			Currency:      currency,
			Status:        status,
			Description:   description,
			FailureReason: failureReason,
		},
		CreatedAt: createdAt,
	}, nil
}

// --- EAV conversion helpers ---

func transferEventDataToAttrs(data transfer.EventData) map[string]string {
	switch d := data.(type) {
	case transfer.TransferCreatedData:
		return map[string]string{
			"from_account_id": d.FromAccountID,
			"to_account_id":   d.ToAccountID,
			"amount":          strconv.FormatInt(d.Amount, 10),
			"currency":        d.Currency,
			"description":     d.Description,
		}
	case transfer.TransferCompletedData:
		return nil
	case transfer.TransferFailedData:
		return map[string]string{
			"reason": d.Reason,
		}
	default:
		return nil
	}
}

func attrsToTransferEventData(eventType transfer.EventType, attrs map[string]string) transfer.EventData {
	switch eventType {
	case transfer.EventTransferCreated:
		amount, _ := strconv.ParseInt(attrs["amount"], 10, 64)
		return transfer.TransferCreatedData{
			FromAccountID: attrs["from_account_id"],
			ToAccountID:   attrs["to_account_id"],
			Amount:        amount,
			Currency:      attrs["currency"],
			Description:   attrs["description"],
		}
	case transfer.EventTransferCompleted:
		return transfer.TransferCompletedData{}
	case transfer.EventTransferFailed:
		return transfer.TransferFailedData{Reason: attrs["reason"]}
	default:
		return nil
	}
}
