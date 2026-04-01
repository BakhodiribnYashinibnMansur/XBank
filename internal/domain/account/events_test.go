package account

import (
	"testing"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
)

func TestNewAccount_RaisesAccountOpenedEvent(t *testing.T) {
	acc, err := NewAccount("user-1", shared.UZS)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := acc.UncommittedEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got: %d", len(events))
	}
	if events[0].Type != EventAccountOpened {
		t.Errorf("expected AccountOpened, got: %s", events[0].Type)
	}
	if acc.Version != 1 {
		t.Errorf("expected version 1, got: %d", acc.Version)
	}
	if acc.ID == "" {
		t.Error("ID should be generated")
	}
}

func TestDeposit_RaisesCreditedEvent(t *testing.T) {
	acc, _ := NewAccount("user-1", shared.UZS)
	acc.ClearUncommittedEvents()

	amount, _ := shared.NewMoney(500000, shared.UZS)
	acc.Deposit(amount)

	events := acc.UncommittedEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got: %d", len(events))
	}
	if events[0].Type != EventCredited {
		t.Errorf("expected Credited, got: %s", events[0].Type)
	}
	if acc.Balance.Amount != 500000 {
		t.Errorf("expected balance 500000, got: %d", acc.Balance.Amount)
	}
}

func TestWithdraw_RaisesDebitedEvent(t *testing.T) {
	acc, _ := NewAccount("user-1", shared.UZS)
	deposit, _ := shared.NewMoney(1000000, shared.UZS)
	acc.Deposit(deposit)
	acc.ClearUncommittedEvents()

	withdraw, _ := shared.NewMoney(300000, shared.UZS)
	acc.Withdraw(withdraw)

	events := acc.UncommittedEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got: %d", len(events))
	}
	if events[0].Type != EventDebited {
		t.Errorf("expected Debited, got: %s", events[0].Type)
	}
	if acc.Balance.Amount != 700000 {
		t.Errorf("expected balance 700000, got: %d", acc.Balance.Amount)
	}
}

func TestLoadFromHistory(t *testing.T) {
	// Simulate events from the event store
	events := []Event{
		{AggregateID: "acc-1", Type: EventAccountOpened, Data: AccountOpenedData{
			UserID: "user-1", AccountNumber: "1234567890123456", Currency: "UZS",
		}, Version: 1, OccurredAt: fixedTime()},
		{AggregateID: "acc-1", Type: EventCredited, Data: CreditedData{
			Amount: 1000000, Currency: "UZS",
		}, Version: 2, OccurredAt: fixedTime()},
		{AggregateID: "acc-1", Type: EventDebited, Data: DebitedData{
			Amount: 300000, Currency: "UZS",
		}, Version: 3, OccurredAt: fixedTime()},
		{AggregateID: "acc-1", Type: EventFrozen, Data: FrozenData{},
			Version: 4, OccurredAt: fixedTime()},
		{AggregateID: "acc-1", Type: EventUnfrozen, Data: UnfrozenData{},
			Version: 5, OccurredAt: fixedTime()},
	}

	acc := LoadFromHistory(events)

	if acc.ID != "acc-1" {
		t.Errorf("expected ID acc-1, got: %s", acc.ID)
	}
	if acc.UserID != "user-1" {
		t.Errorf("expected UserID user-1, got: %s", acc.UserID)
	}
	if acc.Balance.Amount != 700000 {
		t.Errorf("expected balance 700000, got: %d", acc.Balance.Amount)
	}
	if acc.Status != StatusActive {
		t.Errorf("expected ACTIVE, got: %s", acc.Status)
	}
	if acc.Version != 5 {
		t.Errorf("expected version 5, got: %d", acc.Version)
	}
}

func TestLoadFromSnapshot(t *testing.T) {
	snap := SnapshotState{
		UserID:        "user-1",
		AccountNumber: "1234567890123456",
		Balance:       700000,
		Currency:      "UZS",
		Status:        "ACTIVE",
	}

	// Events after snapshot
	events := []Event{
		{AggregateID: "acc-1", Type: EventCredited, Data: CreditedData{
			Amount: 100000, Currency: "UZS",
		}, Version: 6, OccurredAt: fixedTime()},
	}

	acc := LoadFromSnapshot(snap, 5, events)

	if acc.Balance.Amount != 800000 {
		t.Errorf("expected balance 800000, got: %d", acc.Balance.Amount)
	}
	if acc.Version != 6 {
		t.Errorf("expected version 6, got: %d", acc.Version)
	}
}

func TestClearUncommittedEvents(t *testing.T) {
	acc, _ := NewAccount("user-1", shared.UZS)
	if len(acc.UncommittedEvents()) != 1 {
		t.Fatal("should have 1 uncommitted event")
	}

	acc.ClearUncommittedEvents()
	if len(acc.UncommittedEvents()) != 0 {
		t.Error("should have 0 uncommitted events after clear")
	}
}

func TestToSnapshotState(t *testing.T) {
	acc, _ := NewAccount("user-1", shared.UZS)
	deposit, _ := shared.NewMoney(500000, shared.UZS)
	acc.Deposit(deposit)

	snap := acc.ToSnapshotState()
	if snap.Balance != 500000 {
		t.Errorf("expected 500000, got: %d", snap.Balance)
	}
	if snap.Currency != "UZS" {
		t.Errorf("expected UZS, got: %s", snap.Currency)
	}
}

func fixedTime() time.Time {
	return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
}
