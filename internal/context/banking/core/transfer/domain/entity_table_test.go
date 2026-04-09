package domain

import (
	"testing"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

func TestNewTransfer_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		fromAccountID string
		toAccountID   string
		amount        domain.Money
		description   string
		wantErr       error
	}{
		{
			name:          "valid transfer",
			fromAccountID: "acc-1",
			toAccountID:   "acc-2",
			amount:        domain.Money{Amount: 500000, Currency: domain.UZS},
			description:   "salary payment",
			wantErr:       nil,
		},
		{
			name:          "minimal amount",
			fromAccountID: "acc-1",
			toAccountID:   "acc-2",
			amount:        domain.Money{Amount: 1, Currency: domain.UZS},
			description:   "",
			wantErr:       nil,
		},
		{
			name:          "same account",
			fromAccountID: "acc-1",
			toAccountID:   "acc-1",
			amount:        domain.Money{Amount: 500000, Currency: domain.UZS},
			description:   "",
			wantErr:       ErrSameAccount,
		},
		{
			name:          "zero amount",
			fromAccountID: "acc-1",
			toAccountID:   "acc-2",
			amount:        domain.Money{Amount: 0, Currency: domain.UZS},
			description:   "",
			wantErr:       ErrInvalidAmount,
		},
		{
			name:          "negative amount",
			fromAccountID: "acc-1",
			toAccountID:   "acc-2",
			amount:        domain.Money{Amount: -100, Currency: domain.UZS},
			description:   "",
			wantErr:       ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, err := NewTransfer(tt.fromAccountID, tt.toAccountID, tt.amount, tt.description)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tr.FromAccountID != tt.fromAccountID {
				t.Errorf("expected from %s, got %s", tt.fromAccountID, tr.FromAccountID)
			}
			if tr.ToAccountID != tt.toAccountID {
				t.Errorf("expected to %s, got %s", tt.toAccountID, tr.ToAccountID)
			}
			if tr.Amount.Amount != tt.amount.Amount {
				t.Errorf("expected amount %d, got %d", tt.amount.Amount, tr.Amount.Amount)
			}
			if tr.Status != StatusPending {
				t.Errorf("expected PENDING, got %s", tr.Status)
			}
			if tr.ID == "" {
				t.Error("ID should be generated")
			}
			if tr.Version != 1 {
				t.Errorf("expected version 1, got %d", tr.Version)
			}
		})
	}
}

func TestTransfer_EventLifecycle(t *testing.T) {
	amount, _ := domain.NewMoney(500000, domain.UZS)

	tests := []struct {
		name          string
		action        func(tr *Transfer)
		wantStatus    Status
		wantVersion   int
		wantEventType EventType
		wantReason    string
	}{
		{
			name:          "complete transfer",
			action:        func(tr *Transfer) { tr.Complete() },
			wantStatus:    StatusCompleted,
			wantVersion:   2,
			wantEventType: EventTransferCompleted,
		},
		{
			name:          "fail transfer with reason",
			action:        func(tr *Transfer) { tr.Fail("insufficient funds") },
			wantStatus:    StatusFailed,
			wantVersion:   2,
			wantEventType: EventTransferFailed,
			wantReason:    "insufficient funds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr, _ := NewTransfer("acc-1", "acc-2", amount, "test")
			tr.ClearUncommittedEvents()

			tt.action(tr)

			if tr.Status != tt.wantStatus {
				t.Errorf("expected status %s, got %s", tt.wantStatus, tr.Status)
			}
			if tr.Version != tt.wantVersion {
				t.Errorf("expected version %d, got %d", tt.wantVersion, tr.Version)
			}

			events := tr.UncommittedEvents()
			if len(events) != 1 {
				t.Fatalf("expected 1 event, got %d", len(events))
			}
			if events[0].Type != tt.wantEventType {
				t.Errorf("expected event type %s, got %s", tt.wantEventType, events[0].Type)
			}
			if tt.wantReason != "" && tr.FailureReason != tt.wantReason {
				t.Errorf("expected reason %q, got %q", tt.wantReason, tr.FailureReason)
			}
		})
	}
}

func TestTransfer_LoadFromHistory(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		events     []Event
		wantStatus Status
		wantAmount int64
	}{
		{
			name: "pending transfer",
			events: []Event{
				{AggregateID: "tx-1", Type: EventTransferCreated, Data: TransferCreatedData{
					FromAccountID: "acc-1", ToAccountID: "acc-2", Amount: 100000, Currency: "UZS", Description: "test",
				}, Version: 1, OccurredAt: now},
			},
			wantStatus: StatusPending,
			wantAmount: 100000,
		},
		{
			name: "completed transfer",
			events: []Event{
				{AggregateID: "tx-1", Type: EventTransferCreated, Data: TransferCreatedData{
					FromAccountID: "acc-1", ToAccountID: "acc-2", Amount: 200000, Currency: "UZS",
				}, Version: 1, OccurredAt: now},
				{AggregateID: "tx-1", Type: EventTransferCompleted, Data: TransferCompletedData{},
					Version: 2, OccurredAt: now},
			},
			wantStatus: StatusCompleted,
			wantAmount: 200000,
		},
		{
			name: "failed transfer",
			events: []Event{
				{AggregateID: "tx-1", Type: EventTransferCreated, Data: TransferCreatedData{
					FromAccountID: "acc-1", ToAccountID: "acc-2", Amount: 300000, Currency: "UZS",
				}, Version: 1, OccurredAt: now},
				{AggregateID: "tx-1", Type: EventTransferFailed, Data: TransferFailedData{
					Reason: "account frozen",
				}, Version: 2, OccurredAt: now},
			},
			wantStatus: StatusFailed,
			wantAmount: 300000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := LoadFromHistory(tt.events)
			if tr.Status != tt.wantStatus {
				t.Errorf("expected status %s, got %s", tt.wantStatus, tr.Status)
			}
			if tr.Amount.Amount != tt.wantAmount {
				t.Errorf("expected amount %d, got %d", tt.wantAmount, tr.Amount.Amount)
			}
			if tr.ID != "tx-1" {
				t.Errorf("expected ID tx-1, got %s", tr.ID)
			}
		})
	}
}

func TestTransfer_SnapshotRoundTrip(t *testing.T) {
	amount, _ := domain.NewMoney(750000, domain.UZS)
	tr, _ := NewTransfer("acc-1", "acc-2", amount, "snapshot test")

	snap := tr.ToSnapshotState()

	if snap.FromAccountID != "acc-1" {
		t.Errorf("expected from acc-1, got %s", snap.FromAccountID)
	}
	if snap.ToAccountID != "acc-2" {
		t.Errorf("expected to acc-2, got %s", snap.ToAccountID)
	}
	if snap.Amount != 750000 {
		t.Errorf("expected amount 750000, got %d", snap.Amount)
	}
	if snap.Currency != "UZS" {
		t.Errorf("expected UZS, got %s", snap.Currency)
	}
	if snap.Status != "PENDING" {
		t.Errorf("expected PENDING, got %s", snap.Status)
	}
	if snap.Description != "snapshot test" {
		t.Errorf("expected description 'snapshot test', got %s", snap.Description)
	}

	// Restore from snapshot + complete event
	now := time.Now()
	restored := LoadFromSnapshot(snap, 1, []Event{
		{AggregateID: "tx-1", Type: EventTransferCompleted, Data: TransferCompletedData{},
			Version: 2, OccurredAt: now},
	})

	if restored.Status != StatusCompleted {
		t.Errorf("expected COMPLETED after snapshot+event, got %s", restored.Status)
	}
	if restored.Amount.Amount != 750000 {
		t.Errorf("expected amount 750000, got %d", restored.Amount.Amount)
	}
	if restored.Version != 2 {
		t.Errorf("expected version 2, got %d", restored.Version)
	}
}

func TestTransfer_UncommittedEvents(t *testing.T) {
	amount, _ := domain.NewMoney(100000, domain.UZS)
	tr, _ := NewTransfer("acc-1", "acc-2", amount, "test")

	events := tr.UncommittedEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 uncommitted event, got %d", len(events))
	}
	if events[0].Type != EventTransferCreated {
		t.Errorf("expected TransferCreated, got %s", events[0].Type)
	}

	tr.ClearUncommittedEvents()
	if len(tr.UncommittedEvents()) != 0 {
		t.Error("should have 0 uncommitted events after clear")
	}

	tr.Complete()
	events = tr.UncommittedEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event after Complete, got %d", len(events))
	}
	if events[0].Type != EventTransferCompleted {
		t.Errorf("expected TransferCompleted, got %s", events[0].Type)
	}
}
