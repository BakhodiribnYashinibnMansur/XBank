package postgres

import (
	"context"
	"strconv"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/transfer"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransferEventRepository struct {
	pool *pgxpool.Pool
}

func NewTransferEventRepository(pool *pgxpool.Pool) *TransferEventRepository {
	return &TransferEventRepository{pool: pool}
}

// Append - persist transfer events as EAV rows
func (r *TransferEventRepository) Append(ctx context.Context, aggregateID string, expectedVersion int, events []transfer.Event) error {
	db := ExtractDBTX(ctx, r.pool)

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
				return apperror.ErrConcurrencyConflict
			}
		}
	}
	return nil
}

// LoadEvents - load all transfer events, reconstruct from EAV rows
func (r *TransferEventRepository) LoadEvents(ctx context.Context, aggregateID string) ([]transfer.Event, error) {
	db := ExtractDBTX(ctx, r.pool)

	rows, err := db.Query(ctx,
		`SELECT aggregate_id, event_type, version, attr_key, attr_value, occurred_at
		 FROM transfer_events
		 WHERE aggregate_id = $1
		 ORDER BY version ASC, attr_key ASC`,
		aggregateID,
	)
	if err != nil {
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

	return events, nil
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
