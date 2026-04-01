package postgres

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/account"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountEventRepository struct {
	pool *pgxpool.Pool
}

func NewAccountEventRepository(pool *pgxpool.Pool) *AccountEventRepository {
	return &AccountEventRepository{pool: pool}
}

// Append - EAV: each event field = separate row
func (r *AccountEventRepository) Append(ctx context.Context, aggregateID string, expectedVersion int, events []account.Event) error {
	db := ExtractDBTX(ctx, r.pool)

	for _, e := range events {
		attrs := eventDataToAttrs(e.Data)

		// Events with no data (Frozen, Unfrozen, Closed) get a marker row
		if len(attrs) == 0 {
			attrs = map[string]string{"_": "_"}
		}

		for key, val := range attrs {
			_, err := db.Exec(ctx,
				`INSERT INTO account_events (aggregate_id, event_type, version, attr_key, attr_value, occurred_at)
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

// LoadEvents - load all events, reconstruct from EAV rows
func (r *AccountEventRepository) LoadEvents(ctx context.Context, aggregateID string) ([]account.Event, error) {
	return r.loadEvents(ctx, aggregateID, 0)
}

// LoadEventsFromVersion - load events after a given version
func (r *AccountEventRepository) LoadEventsFromVersion(ctx context.Context, aggregateID string, fromVersion int) ([]account.Event, error) {
	return r.loadEvents(ctx, aggregateID, fromVersion)
}

func (r *AccountEventRepository) loadEvents(ctx context.Context, aggregateID string, fromVersion int) ([]account.Event, error) {
	db := ExtractDBTX(ctx, r.pool)

	rows, err := db.Query(ctx,
		`SELECT aggregate_id, event_type, version, attr_key, attr_value, occurred_at
		 FROM account_events
		 WHERE aggregate_id = $1 AND version > $2 AND event_type != 'Snapshot'
		 ORDER BY version ASC, attr_key ASC`,
		aggregateID, fromVersion,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Group rows by (aggregate_id, version) into events
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
	var events []account.Event
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
		eventType := account.EventType(first.eventType)
		data := attrsToEventData(eventType, attrs)

		events = append(events, account.Event{
			AggregateID: first.aggregateID,
			Type:        eventType,
			Data:        data,
			Version:     first.version,
			OccurredAt:  first.occurredAt,
		})
	}

	return events, nil
}

// SaveSnapshot - store as event_type = 'Snapshot'
func (r *AccountEventRepository) SaveSnapshot(ctx context.Context, snapshot account.Snapshot) error {
	db := ExtractDBTX(ctx, r.pool)

	// Delete old snapshot rows for this aggregate
	_, err := db.Exec(ctx,
		`DELETE FROM account_events WHERE aggregate_id = $1 AND event_type = 'Snapshot'`,
		snapshot.AggregateID,
	)
	if err != nil {
		return err
	}

	// Insert snapshot as EAV rows (version = snapshot version, special event_type)
	attrs := map[string]string{
		"user_id":        snapshot.State.UserID,
		"account_number": snapshot.State.AccountNumber,
		"balance":        strconv.FormatInt(snapshot.State.Balance, 10),
		"currency":       snapshot.State.Currency,
		"status":         snapshot.State.Status,
	}

	for key, val := range attrs {
		_, err := db.Exec(ctx,
			`INSERT INTO account_events (aggregate_id, event_type, version, attr_key, attr_value, occurred_at)
			 VALUES ($1, 'Snapshot', $2, $3, $4, $5)`,
			snapshot.AggregateID, snapshot.Version, key, val, snapshot.CreatedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadSnapshot - load latest snapshot from EAV rows
func (r *AccountEventRepository) LoadSnapshot(ctx context.Context, aggregateID string) (*account.Snapshot, error) {
	db := ExtractDBTX(ctx, r.pool)

	rows, err := db.Query(ctx,
		`SELECT version, attr_key, attr_value, occurred_at
		 FROM account_events
		 WHERE aggregate_id = $1 AND event_type = 'Snapshot'`,
		aggregateID,
	)
	if err != nil {
		return nil, nil
	}
	defer rows.Close()

	attrs := make(map[string]string)
	var version int
	var createdAt time.Time
	found := false

	for rows.Next() {
		var key, val string
		if err := rows.Scan(&version, &key, &val, &createdAt); err != nil {
			return nil, err
		}
		attrs[key] = val
		found = true
	}

	if !found {
		return nil, nil
	}

	balance, _ := strconv.ParseInt(attrs["balance"], 10, 64)

	return &account.Snapshot{
		AggregateID: aggregateID,
		Version:     version,
		State: account.SnapshotState{
			UserID:        attrs["user_id"],
			AccountNumber: attrs["account_number"],
			Balance:       balance,
			Currency:      attrs["currency"],
			Status:        attrs["status"],
		},
		CreatedAt: createdAt,
	}, nil
}

// --- EAV conversion helpers ---

// eventDataToAttrs converts domain EventData to key-value pairs
func eventDataToAttrs(data account.EventData) map[string]string {
	switch d := data.(type) {
	case account.AccountOpenedData:
		return map[string]string{
			"user_id":        d.UserID,
			"account_number": d.AccountNumber,
			"currency":       d.Currency,
		}
	case account.CreditedData:
		return map[string]string{
			"amount":   strconv.FormatInt(d.Amount, 10),
			"currency": d.Currency,
		}
	case account.DebitedData:
		return map[string]string{
			"amount":   strconv.FormatInt(d.Amount, 10),
			"currency": d.Currency,
		}
	case account.FrozenData, account.UnfrozenData, account.ClosedData:
		return nil
	default:
		return nil
	}
}

// attrsToEventData converts EAV key-value pairs back to domain EventData
func attrsToEventData(eventType account.EventType, attrs map[string]string) account.EventData {
	switch eventType {
	case account.EventAccountOpened:
		return account.AccountOpenedData{
			UserID:        attrs["user_id"],
			AccountNumber: attrs["account_number"],
			Currency:      attrs["currency"],
		}
	case account.EventCredited:
		amount, _ := strconv.ParseInt(attrs["amount"], 10, 64)
		return account.CreditedData{Amount: amount, Currency: attrs["currency"]}
	case account.EventDebited:
		amount, _ := strconv.ParseInt(attrs["amount"], 10, 64)
		return account.DebitedData{Amount: amount, Currency: attrs["currency"]}
	case account.EventFrozen:
		return account.FrozenData{}
	case account.EventUnfrozen:
		return account.UnfrozenData{}
	case account.EventClosed:
		return account.ClosedData{}
	default:
		return nil
	}
}

func init() {
	// Ensure we don't accidentally use fmt (silence import)
	_ = fmt.Sprintf
}
