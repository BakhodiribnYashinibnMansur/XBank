package postgres

import (
	"context"
	"fmt"
	"strconv"
	"time"

	account "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/domain"
	sharedpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EventRepo struct {
	pool *pgxpool.Pool
}

func NewEventRepo(pool *pgxpool.Pool) *EventRepo {
	return &EventRepo{pool: pool}
}

// Append - EAV: each event field = separate row
func (r *EventRepo) Append(ctx context.Context, aggregateID string, expectedVersion int, events []account.Event) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)

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
				metrics.ObserveQuery("AccountEventRepo.Append", start, err)
				return apperror.ErrConcurrencyConflict
			}
		}
	}
	metrics.ObserveQuery("AccountEventRepo.Append", start, nil)
	return nil
}

// LoadEvents - load all events, reconstruct from EAV rows
func (r *EventRepo) LoadEvents(ctx context.Context, aggregateID string) ([]account.Event, error) {
	return r.loadEvents(ctx, aggregateID, 0)
}

// LoadEventsFromVersion - load events after a given version
func (r *EventRepo) LoadEventsFromVersion(ctx context.Context, aggregateID string, fromVersion int) ([]account.Event, error) {
	return r.loadEvents(ctx, aggregateID, fromVersion)
}

func (r *EventRepo) loadEvents(ctx context.Context, aggregateID string, fromVersion int) ([]account.Event, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)

	rows, err := db.Query(ctx,
		`SELECT aggregate_id, event_type, version, attr_key, attr_value, occurred_at
		 FROM account_events
		 WHERE aggregate_id = $1 AND version > $2 AND event_type != 'Snapshot'
		 ORDER BY version ASC, attr_key ASC`,
		aggregateID, fromVersion,
	)
	if err != nil {
		metrics.ObserveQuery("AccountEventRepo.LoadEvents", start, err)
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

	metrics.ObserveQuery("AccountEventRepo.LoadEvents", start, nil)
	return events, nil
}

// SaveSnapshot - upsert into dedicated account_snapshots table
func (r *EventRepo) SaveSnapshot(ctx context.Context, snapshot account.Snapshot) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)

	_, err := db.Exec(ctx,
		`INSERT INTO account_snapshots (aggregate_id, version, user_id, account_number, balance, currency, status, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (aggregate_id)
		 DO UPDATE SET version = EXCLUDED.version,
		               user_id = EXCLUDED.user_id,
		               account_number = EXCLUDED.account_number,
		               balance = EXCLUDED.balance,
		               currency = EXCLUDED.currency,
		               status = EXCLUDED.status,
		               created_at = EXCLUDED.created_at`,
		snapshot.AggregateID, snapshot.Version,
		snapshot.State.UserID, snapshot.State.AccountNumber,
		snapshot.State.Balance, snapshot.State.Currency, snapshot.State.Status,
		snapshot.CreatedAt,
	)
	if err != nil {
		metrics.ObserveQuery("AccountEventRepo.SaveSnapshot", start, err)
		return err
	}

	metrics.ObserveQuery("AccountEventRepo.SaveSnapshot", start, nil)
	return nil
}

// LoadSnapshot - load latest snapshot from dedicated table
func (r *EventRepo) LoadSnapshot(ctx context.Context, aggregateID string) (*account.Snapshot, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)

	var snap account.Snapshot
	var userID, accountNumber, currency, status string
	var balance int64
	var version int
	var createdAt time.Time

	err := db.QueryRow(ctx,
		`SELECT version, user_id, account_number, balance, currency, status, created_at
		 FROM account_snapshots
		 WHERE aggregate_id = $1`,
		aggregateID,
	).Scan(&version, &userID, &accountNumber, &balance, &currency, &status, &createdAt)

	if err != nil {
		metrics.ObserveQuery("AccountEventRepo.LoadSnapshot", start, nil)
		return nil, nil // not found = no snapshot
	}

	snap = account.Snapshot{
		AggregateID: aggregateID,
		Version:     version,
		State: account.SnapshotState{
			UserID:        userID,
			AccountNumber: accountNumber,
			Balance:       balance,
			Currency:      currency,
			Status:        status,
		},
		CreatedAt: createdAt,
	}

	metrics.ObserveQuery("AccountEventRepo.LoadSnapshot", start, nil)
	return &snap, nil
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
