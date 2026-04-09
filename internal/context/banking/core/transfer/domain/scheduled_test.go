package domain

import (
	"testing"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

func TestNewScheduledTransfer_TableDriven(t *testing.T) {
	futureTime := time.Now().Add(24 * time.Hour)
	pastTime := time.Now().Add(-1 * time.Hour)

	tests := []struct {
		name          string
		userID        string
		fromAccountID string
		toAccountID   string
		amount        domain.Money
		description   string
		executeAt     time.Time
		wantErr       bool
	}{
		{
			name:          "valid scheduled transfer",
			userID:        "user-1",
			fromAccountID: "acc-1",
			toAccountID:   "acc-2",
			amount:        domain.Money{Amount: 500000, Currency: domain.UZS},
			description:   "salary",
			executeAt:     futureTime,
			wantErr:       false,
		},
		{
			name:          "same account",
			userID:        "user-1",
			fromAccountID: "acc-1",
			toAccountID:   "acc-1",
			amount:        domain.Money{Amount: 500000, Currency: domain.UZS},
			description:   "",
			executeAt:     futureTime,
			wantErr:       true,
		},
		{
			name:          "zero amount",
			userID:        "user-1",
			fromAccountID: "acc-1",
			toAccountID:   "acc-2",
			amount:        domain.Money{Amount: 0, Currency: domain.UZS},
			description:   "",
			executeAt:     futureTime,
			wantErr:       true,
		},
		{
			name:          "negative amount",
			userID:        "user-1",
			fromAccountID: "acc-1",
			toAccountID:   "acc-2",
			amount:        domain.Money{Amount: -100, Currency: domain.UZS},
			description:   "",
			executeAt:     futureTime,
			wantErr:       true,
		},
		{
			name:          "past execute time",
			userID:        "user-1",
			fromAccountID: "acc-1",
			toAccountID:   "acc-2",
			amount:        domain.Money{Amount: 100000, Currency: domain.UZS},
			description:   "",
			executeAt:     pastTime,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st, err := NewScheduledTransfer(tt.userID, tt.fromAccountID, tt.toAccountID, tt.amount, tt.description, tt.executeAt)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if st.ID == "" {
				t.Error("ID should be generated")
			}
			if st.Status != ScheduledPending {
				t.Errorf("expected PENDING, got %s", st.Status)
			}
			if st.UserID != tt.userID {
				t.Errorf("expected userID %s, got %s", tt.userID, st.UserID)
			}
			if st.FromAccountID != tt.fromAccountID {
				t.Errorf("expected from %s, got %s", tt.fromAccountID, st.FromAccountID)
			}
			if st.ToAccountID != tt.toAccountID {
				t.Errorf("expected to %s, got %s", tt.toAccountID, st.ToAccountID)
			}
		})
	}
}

func TestScheduledTransfer_Cancel_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *ScheduledTransfer
		wantErr bool
	}{
		{
			name: "cancel pending",
			setup: func() *ScheduledTransfer {
				st, _ := NewScheduledTransfer("user-1", "acc-1", "acc-2",
					domain.Money{Amount: 100000, Currency: domain.UZS}, "", time.Now().Add(time.Hour))
				return st
			},
			wantErr: false,
		},
		{
			name: "cancel executed",
			setup: func() *ScheduledTransfer {
				st, _ := NewScheduledTransfer("user-1", "acc-1", "acc-2",
					domain.Money{Amount: 100000, Currency: domain.UZS}, "", time.Now().Add(time.Hour))
				st.MarkExecuted("tx-123")
				return st
			},
			wantErr: true,
		},
		{
			name: "cancel failed",
			setup: func() *ScheduledTransfer {
				st, _ := NewScheduledTransfer("user-1", "acc-1", "acc-2",
					domain.Money{Amount: 100000, Currency: domain.UZS}, "", time.Now().Add(time.Hour))
				st.MarkFailed("some reason")
				return st
			},
			wantErr: true,
		},
		{
			name: "cancel already cancelled",
			setup: func() *ScheduledTransfer {
				st, _ := NewScheduledTransfer("user-1", "acc-1", "acc-2",
					domain.Money{Amount: 100000, Currency: domain.UZS}, "", time.Now().Add(time.Hour))
				st.Cancel()
				return st
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := tt.setup()
			err := st.Cancel()
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if st.Status != ScheduledCancelled {
				t.Errorf("expected CANCELLED, got %s", st.Status)
			}
		})
	}
}

func TestScheduledTransfer_MarkExecuted(t *testing.T) {
	st, _ := NewScheduledTransfer("user-1", "acc-1", "acc-2",
		domain.Money{Amount: 100000, Currency: domain.UZS}, "test", time.Now().Add(time.Hour))

	st.MarkExecuted("tx-456")

	if st.Status != ScheduledExecuted {
		t.Errorf("expected EXECUTED, got %s", st.Status)
	}
	if st.TransferID != "tx-456" {
		t.Errorf("expected transferID tx-456, got %s", st.TransferID)
	}
	if st.ExecutedAt == nil {
		t.Error("ExecutedAt should be set")
	}
}

func TestScheduledTransfer_MarkFailed(t *testing.T) {
	st, _ := NewScheduledTransfer("user-1", "acc-1", "acc-2",
		domain.Money{Amount: 100000, Currency: domain.UZS}, "test", time.Now().Add(time.Hour))

	st.MarkFailed("insufficient funds")

	if st.Status != ScheduledFailed {
		t.Errorf("expected FAILED, got %s", st.Status)
	}
	if st.FailureReason != "insufficient funds" {
		t.Errorf("expected reason 'insufficient funds', got %s", st.FailureReason)
	}
	if st.ExecutedAt == nil {
		t.Error("ExecutedAt should be set")
	}
}
