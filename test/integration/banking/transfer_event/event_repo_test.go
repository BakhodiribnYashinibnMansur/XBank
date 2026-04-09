package transfer_event_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/infrastructure/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/test/integration/common/setup"
)

var pgc *setup.PostgresContainer

func TestMain(m *testing.M) {
	if os.Getenv("INTEGRATION") == "" {
		fmt.Println("skipping integration tests (set INTEGRATION=1 to run)")
		os.Exit(0)
	}

	var code int
	func() {
		t := &testing.T{}
		pgc = setup.MustPostgres(t)
		defer pgc.Teardown()
		code = m.Run()
	}()
	os.Exit(code)
}

func truncate(t *testing.T) {
	t.Helper()
	setup.TruncateAll(t, pgc.Pool)
}

func TestEventRepo_AppendAndLoadEvents(t *testing.T) {
	truncate(t)
	repo := pg.NewEventRepo(pgc.Pool)
	ctx := context.Background()

	aggregateID := "a0000000-0000-4000-8000-000000000001"
	events := []domain.Event{
		{
			AggregateID: aggregateID,
			Type:        domain.EventTransferCreated,
			Data: domain.TransferCreatedData{
				FromAccountID: "from-acc-id",
				ToAccountID:   "to-acc-id",
				Amount:        50000,
				Currency:      "UZS",
				Description:   "test transfer",
			},
			Version:    1,
			OccurredAt: time.Now(),
		},
	}

	if err := repo.Append(ctx, aggregateID, events); err != nil {
		t.Fatalf("Append: %v", err)
	}

	loaded, err := repo.LoadEvents(ctx, aggregateID)
	if err != nil {
		t.Fatalf("LoadEvents: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("got %d events, want 1", len(loaded))
	}

	e := loaded[0]
	if e.AggregateID != aggregateID {
		t.Errorf("AggregateID = %q, want %q", e.AggregateID, aggregateID)
	}
	if e.Type != domain.EventTransferCreated {
		t.Errorf("Type = %q, want %q", e.Type, domain.EventTransferCreated)
	}
	if e.Version != 1 {
		t.Errorf("Version = %d, want 1", e.Version)
	}

	data, ok := e.Data.(domain.TransferCreatedData)
	if !ok {
		t.Fatal("expected TransferCreatedData")
	}
	if data.Amount != 50000 {
		t.Errorf("Amount = %d, want 50000", data.Amount)
	}
	if data.Description != "test transfer" {
		t.Errorf("Description = %q, want %q", data.Description, "test transfer")
	}
}

func TestEventRepo_AppendMultipleAndLoadFromVersion(t *testing.T) {
	truncate(t)
	repo := pg.NewEventRepo(pgc.Pool)
	ctx := context.Background()

	aggregateID := "a0000000-0000-4000-8000-000000000002"

	// Append event 1: TransferCreated
	events1 := []domain.Event{
		{
			AggregateID: aggregateID,
			Type:        domain.EventTransferCreated,
			Data: domain.TransferCreatedData{
				FromAccountID: "from-acc",
				ToAccountID:   "to-acc",
				Amount:        10000,
				Currency:      "UZS",
				Description:   "multi test",
			},
			Version:    1,
			OccurredAt: time.Now(),
		},
	}
	if err := repo.Append(ctx, aggregateID, events1); err != nil {
		t.Fatalf("Append events1: %v", err)
	}

	// Append event 2: TransferCompleted
	events2 := []domain.Event{
		{
			AggregateID: aggregateID,
			Type:        domain.EventTransferCompleted,
			Data:        domain.TransferCompletedData{},
			Version:     2,
			OccurredAt:  time.Now(),
		},
	}
	if err := repo.Append(ctx, aggregateID, events2); err != nil {
		t.Fatalf("Append events2: %v", err)
	}

	// Load all
	all, err := repo.LoadEvents(ctx, aggregateID)
	if err != nil {
		t.Fatalf("LoadEvents: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("got %d events, want 2", len(all))
	}

	// Load from version 1 (should return only version 2)
	fromV1, err := repo.LoadEventsFromVersion(ctx, aggregateID, 1)
	if err != nil {
		t.Fatalf("LoadEventsFromVersion: %v", err)
	}
	if len(fromV1) != 1 {
		t.Fatalf("got %d events from version 1, want 1", len(fromV1))
	}
	if fromV1[0].Version != 2 {
		t.Errorf("Version = %d, want 2", fromV1[0].Version)
	}
	if fromV1[0].Type != domain.EventTransferCompleted {
		t.Errorf("Type = %q, want %q", fromV1[0].Type, domain.EventTransferCompleted)
	}
}

func TestEventRepo_TransferFailedEvent(t *testing.T) {
	truncate(t)
	repo := pg.NewEventRepo(pgc.Pool)
	ctx := context.Background()

	aggregateID := "a0000000-0000-4000-8000-000000000003"

	events := []domain.Event{
		{
			AggregateID: aggregateID,
			Type:        domain.EventTransferFailed,
			Data:        domain.TransferFailedData{Reason: "insufficient funds"},
			Version:     1,
			OccurredAt:  time.Now(),
		},
	}
	if err := repo.Append(ctx, aggregateID, events); err != nil {
		t.Fatalf("Append: %v", err)
	}

	loaded, err := repo.LoadEvents(ctx, aggregateID)
	if err != nil {
		t.Fatalf("LoadEvents: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("got %d events, want 1", len(loaded))
	}

	data, ok := loaded[0].Data.(domain.TransferFailedData)
	if !ok {
		t.Fatal("expected TransferFailedData")
	}
	if data.Reason != "insufficient funds" {
		t.Errorf("Reason = %q, want %q", data.Reason, "insufficient funds")
	}
}

func TestEventRepo_SaveAndLoadSnapshot(t *testing.T) {
	truncate(t)
	repo := pg.NewEventRepo(pgc.Pool)
	ctx := context.Background()

	aggregateID := "a0000000-0000-4000-8000-000000000004"
	fromAccUUID := "f0000000-0000-4000-8000-000000000001"
	toAccUUID := "f0000000-0000-4000-8000-000000000002"
	snap := domain.Snapshot{
		AggregateID: aggregateID,
		Version:     3,
		State: domain.SnapshotState{
			FromAccountID: fromAccUUID,
			ToAccountID:   toAccUUID,
			Amount:        75000,
			Currency:      "UZS",
			Status:        "COMPLETED",
			Description:   "snapshot test",
			FailureReason: "",
		},
		CreatedAt: time.Now(),
	}

	if err := repo.SaveSnapshot(ctx, snap); err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}

	loaded, err := repo.LoadSnapshot(ctx, aggregateID)
	if err != nil {
		t.Fatalf("LoadSnapshot: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected snapshot, got nil")
	}
	if loaded.Version != 3 {
		t.Errorf("Version = %d, want 3", loaded.Version)
	}
	if loaded.State.Amount != 75000 {
		t.Errorf("Amount = %d, want 75000", loaded.State.Amount)
	}
	if loaded.State.Status != "COMPLETED" {
		t.Errorf("Status = %q, want COMPLETED", loaded.State.Status)
	}
}

func TestEventRepo_SaveSnapshot_Upsert(t *testing.T) {
	truncate(t)
	repo := pg.NewEventRepo(pgc.Pool)
	ctx := context.Background()

	aggregateID := "a0000000-0000-4000-8000-000000000005"
	fromUUID := "f0000000-0000-4000-8000-000000000003"
	toUUID := "f0000000-0000-4000-8000-000000000004"
	snap1 := domain.Snapshot{
		AggregateID: aggregateID,
		Version:     1,
		State: domain.SnapshotState{
			FromAccountID: fromUUID,
			ToAccountID:   toUUID,
			Amount:        10000,
			Currency:      "UZS",
			Status:        "PENDING",
		},
		CreatedAt: time.Now(),
	}
	repo.SaveSnapshot(ctx, snap1)

	snap2 := domain.Snapshot{
		AggregateID: aggregateID,
		Version:     5,
		State: domain.SnapshotState{
			FromAccountID: fromUUID,
			ToAccountID:   toUUID,
			Amount:        10000,
			Currency:      "UZS",
			Status:        "COMPLETED",
		},
		CreatedAt: time.Now(),
	}
	if err := repo.SaveSnapshot(ctx, snap2); err != nil {
		t.Fatalf("SaveSnapshot upsert: %v", err)
	}

	loaded, _ := repo.LoadSnapshot(ctx, aggregateID)
	if loaded.Version != 5 {
		t.Errorf("Version = %d, want 5 after upsert", loaded.Version)
	}
	if loaded.State.Status != "COMPLETED" {
		t.Errorf("Status = %q, want COMPLETED after upsert", loaded.State.Status)
	}
}

func TestEventRepo_LoadSnapshot_NotFound(t *testing.T) {
	truncate(t)
	repo := pg.NewEventRepo(pgc.Pool)
	ctx := context.Background()

	snap, err := repo.LoadSnapshot(ctx, "a0000000-0000-4000-8000-999999999999")
	if err != nil {
		t.Fatalf("LoadSnapshot: %v", err)
	}
	if snap != nil {
		t.Error("expected nil snapshot for nonexistent aggregate")
	}
}

func TestEventRepo_LoadEvents_Empty(t *testing.T) {
	truncate(t)
	repo := pg.NewEventRepo(pgc.Pool)
	ctx := context.Background()

	events, err := repo.LoadEvents(ctx, "a0000000-0000-4000-8000-999999999999")
	if err != nil {
		t.Fatalf("LoadEvents: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("got %d events, want 0", len(events))
	}
}
