package accountevent_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/domain"
	pg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/infrastructure/postgres"
	kerneldomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
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

func TestEventRepo_AppendAndLoad(t *testing.T) {
	truncate(t)
	repo := pg.NewEventRepo(pgc.Pool)
	ctx := context.Background()

	acc, err := domain.NewAccount("user-123", kerneldomain.UZS)
	if err != nil {
		t.Fatalf("NewAccount: %v", err)
	}

	events := acc.UncommittedEvents()
	if len(events) == 0 {
		t.Fatal("expected uncommitted events from NewAccount")
	}

	if err := repo.Append(ctx, acc.ID, events); err != nil {
		t.Fatalf("Append: %v", err)
	}

	loaded, err := repo.LoadEvents(ctx, acc.ID)
	if err != nil {
		t.Fatalf("LoadEvents: %v", err)
	}
	if len(loaded) != 1 {
		t.Fatalf("loaded %d events, want 1", len(loaded))
	}
	if loaded[0].Type != domain.EventAccountOpened {
		t.Errorf("event type = %q, want %q", loaded[0].Type, domain.EventAccountOpened)
	}
}

func TestEventRepo_AppendMultipleEvents(t *testing.T) {
	truncate(t)
	repo := pg.NewEventRepo(pgc.Pool)
	ctx := context.Background()

	acc, _ := domain.NewAccount("user-456", kerneldomain.UZS)
	repo.Append(ctx, acc.ID, acc.UncommittedEvents())
	acc.ClearUncommittedEvents()

	acc.Deposit(kerneldomain.Money{Amount: 50_000, Currency: kerneldomain.UZS})
	acc.Deposit(kerneldomain.Money{Amount: 30_000, Currency: kerneldomain.UZS})
	acc.Withdraw(kerneldomain.Money{Amount: 10_000, Currency: kerneldomain.UZS})

	if err := repo.Append(ctx, acc.ID, acc.UncommittedEvents()); err != nil {
		t.Fatalf("Append: %v", err)
	}

	loaded, err := repo.LoadEvents(ctx, acc.ID)
	if err != nil {
		t.Fatalf("LoadEvents: %v", err)
	}
	if len(loaded) != 4 {
		t.Fatalf("loaded %d events, want 4", len(loaded))
	}

	expectedTypes := []domain.EventType{
		domain.EventAccountOpened,
		domain.EventCredited,
		domain.EventCredited,
		domain.EventDebited,
	}
	for i, et := range expectedTypes {
		if loaded[i].Type != et {
			t.Errorf("event[%d].Type = %q, want %q", i, loaded[i].Type, et)
		}
	}
}

func TestEventRepo_LoadEventsFromVersion(t *testing.T) {
	truncate(t)
	repo := pg.NewEventRepo(pgc.Pool)
	ctx := context.Background()

	acc, _ := domain.NewAccount("user-789", kerneldomain.UZS)
	repo.Append(ctx, acc.ID, acc.UncommittedEvents())
	acc.ClearUncommittedEvents()

	acc.Deposit(kerneldomain.Money{Amount: 100_000, Currency: kerneldomain.UZS})
	acc.Deposit(kerneldomain.Money{Amount: 200_000, Currency: kerneldomain.UZS})
	repo.Append(ctx, acc.ID, acc.UncommittedEvents())

	loaded, err := repo.LoadEventsFromVersion(ctx, acc.ID, 1)
	if err != nil {
		t.Fatalf("LoadEventsFromVersion: %v", err)
	}
	if len(loaded) != 2 {
		t.Fatalf("loaded %d events, want 2", len(loaded))
	}
	if loaded[0].Type != domain.EventCredited {
		t.Errorf("event[0].Type = %q, want Credited", loaded[0].Type)
	}
}

func TestEventRepo_SnapshotSaveAndLoad(t *testing.T) {
	truncate(t)
	repo := pg.NewEventRepo(pgc.Pool)
	ctx := context.Background()

	aggID := "a0000000-0000-4000-8000-000000000001"
	snapshot := domain.Snapshot{
		AggregateID: aggID,
		Version:     5,
		State: domain.SnapshotState{
			UserID:        "b0000000-0000-4000-8000-000000000001",
			AccountNumber: "1234567890123456",
			Balance:       500_000,
			Currency:      "UZS",
			Status:        "ACTIVE",
		},
		CreatedAt: time.Now().Truncate(time.Microsecond),
	}

	if err := repo.SaveSnapshot(ctx, snapshot); err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}

	loaded, err := repo.LoadSnapshot(ctx, aggID)
	if err != nil {
		t.Fatalf("LoadSnapshot: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected snapshot, got nil")
	}
	if loaded.Version != 5 {
		t.Errorf("version = %d, want 5", loaded.Version)
	}
	if loaded.State.Balance != 500_000 {
		t.Errorf("balance = %d, want 500000", loaded.State.Balance)
	}
}

func TestEventRepo_SnapshotUpsert(t *testing.T) {
	truncate(t)
	repo := pg.NewEventRepo(pgc.Pool)
	ctx := context.Background()

	aggID := "c0000000-0000-4000-8000-000000000002"
	snap1 := domain.Snapshot{
		AggregateID: aggID,
		Version:     3,
		State: domain.SnapshotState{
			UserID: "d0000000-0000-4000-8000-000000000001", AccountNumber: "1111111111111111",
			Balance: 100, Currency: "UZS", Status: "ACTIVE",
		},
		CreatedAt: time.Now(),
	}
	repo.SaveSnapshot(ctx, snap1)

	snap2 := snap1
	snap2.Version = 10
	snap2.State.Balance = 999_000
	repo.SaveSnapshot(ctx, snap2)

	loaded, _ := repo.LoadSnapshot(ctx, aggID)
	if loaded.Version != 10 {
		t.Errorf("version = %d, want 10 (upsert)", loaded.Version)
	}
	if loaded.State.Balance != 999_000 {
		t.Errorf("balance = %d, want 999000", loaded.State.Balance)
	}
}

func TestEventRepo_LoadSnapshot_NotFound(t *testing.T) {
	truncate(t)
	repo := pg.NewEventRepo(pgc.Pool)
	ctx := context.Background()

	snap, err := repo.LoadSnapshot(ctx, "e0000000-0000-4000-8000-000000000099")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap != nil {
		t.Error("expected nil snapshot for nonexistent aggregate")
	}
}
