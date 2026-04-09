package domain

import (
	"testing"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

func TestNewAccount_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		userID    string
		currency  domain.Currency
		wantErr   error
		wantState func(t *testing.T, acc *Account)
	}{
		{
			name:     "valid UZS account",
			userID:   "user-1",
			currency: domain.UZS,
			wantErr:  nil,
			wantState: func(t *testing.T, acc *Account) {
				if acc.Status != StatusActive {
					t.Errorf("expected ACTIVE, got %s", acc.Status)
				}
				if acc.Balance.Currency != domain.UZS {
					t.Errorf("expected UZS, got %s", acc.Balance.Currency)
				}
				if !acc.Balance.IsZero() {
					t.Error("new account balance should be zero")
				}
				if len(acc.AccountNumber) != 16 {
					t.Errorf("account number should be 16 chars, got %d", len(acc.AccountNumber))
				}
				if acc.ID == "" {
					t.Error("ID should be generated")
				}
				if acc.Version != 1 {
					t.Errorf("expected version 1, got %d", acc.Version)
				}
			},
		},
		{
			name:     "valid USD account",
			userID:   "user-2",
			currency: domain.USD,
			wantErr:  nil,
			wantState: func(t *testing.T, acc *Account) {
				if acc.Balance.Currency != domain.USD {
					t.Errorf("expected USD, got %s", acc.Balance.Currency)
				}
			},
		},
		{
			name:     "valid EUR account",
			userID:   "user-3",
			currency: domain.EUR,
			wantErr:  nil,
			wantState: func(t *testing.T, acc *Account) {
				if acc.Balance.Currency != domain.EUR {
					t.Errorf("expected EUR, got %s", acc.Balance.Currency)
				}
			},
		},
		{
			name:     "empty userID",
			userID:   "",
			currency: domain.UZS,
			wantErr:  ErrMissingUserID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc, err := NewAccount(tt.userID, tt.currency)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantState != nil {
				tt.wantState(t, acc)
			}
		})
	}
}

func TestAccount_Deposit_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *Account
		amount      domain.Money
		wantErr     error
		wantBalance int64
	}{
		{
			name: "deposit to active UZS account",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				return acc
			},
			amount:      domain.Money{Amount: 1000000, Currency: domain.UZS},
			wantErr:     nil,
			wantBalance: 1000000,
		},
		{
			name: "deposit small amount",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				return acc
			},
			amount:      domain.Money{Amount: 1, Currency: domain.UZS},
			wantErr:     nil,
			wantBalance: 1,
		},
		{
			name: "deposit to frozen account",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				acc.Freeze()
				return acc
			},
			amount:  domain.Money{Amount: 100000, Currency: domain.UZS},
			wantErr: ErrAccountFrozen,
		},
		{
			name: "deposit to closed account",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				acc.Close()
				return acc
			},
			amount:  domain.Money{Amount: 100000, Currency: domain.UZS},
			wantErr: ErrAccountClosed,
		},
		{
			name: "deposit with currency mismatch",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				return acc
			},
			amount:  domain.Money{Amount: 100000, Currency: domain.USD},
			wantErr: domain.ErrCurrencyMismatch,
		},
		{
			name: "multiple deposits accumulate",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				acc.Deposit(domain.Money{Amount: 500000, Currency: domain.UZS})
				return acc
			},
			amount:      domain.Money{Amount: 300000, Currency: domain.UZS},
			wantErr:     nil,
			wantBalance: 800000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := tt.setup()
			err := acc.Deposit(tt.amount)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if acc.Balance.Amount != tt.wantBalance {
				t.Errorf("expected balance %d, got %d", tt.wantBalance, acc.Balance.Amount)
			}
		})
	}
}

func TestAccount_Withdraw_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *Account
		amount      domain.Money
		wantErr     error
		wantBalance int64
	}{
		{
			name: "withdraw within balance",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				acc.Deposit(domain.Money{Amount: 1000000, Currency: domain.UZS})
				return acc
			},
			amount:      domain.Money{Amount: 300000, Currency: domain.UZS},
			wantErr:     nil,
			wantBalance: 700000,
		},
		{
			name: "withdraw exact balance",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				acc.Deposit(domain.Money{Amount: 500000, Currency: domain.UZS})
				return acc
			},
			amount:      domain.Money{Amount: 500000, Currency: domain.UZS},
			wantErr:     nil,
			wantBalance: 0,
		},
		{
			name: "withdraw insufficient funds",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				acc.Deposit(domain.Money{Amount: 100000, Currency: domain.UZS})
				return acc
			},
			amount:  domain.Money{Amount: 200000, Currency: domain.UZS},
			wantErr: domain.ErrInsufficientFunds,
		},
		{
			name: "withdraw from empty account",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				return acc
			},
			amount:  domain.Money{Amount: 1, Currency: domain.UZS},
			wantErr: domain.ErrInsufficientFunds,
		},
		{
			name: "withdraw from frozen account",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				acc.Deposit(domain.Money{Amount: 1000000, Currency: domain.UZS})
				acc.Freeze()
				return acc
			},
			amount:  domain.Money{Amount: 100000, Currency: domain.UZS},
			wantErr: ErrAccountFrozen,
		},
		{
			name: "withdraw from closed account",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				acc.Close()
				return acc
			},
			amount:  domain.Money{Amount: 100000, Currency: domain.UZS},
			wantErr: ErrAccountClosed,
		},
		{
			name: "withdraw currency mismatch",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				acc.Deposit(domain.Money{Amount: 1000000, Currency: domain.UZS})
				return acc
			},
			amount:  domain.Money{Amount: 100, Currency: domain.USD},
			wantErr: domain.ErrCurrencyMismatch,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := tt.setup()
			err := acc.Withdraw(tt.amount)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if acc.Balance.Amount != tt.wantBalance {
				t.Errorf("expected balance %d, got %d", tt.wantBalance, acc.Balance.Amount)
			}
		})
	}
}

func TestAccount_Freeze_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Account
		wantErr error
	}{
		{
			name: "freeze active account",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				return acc
			},
			wantErr: nil,
		},
		{
			name: "freeze already frozen account",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				acc.Freeze()
				return acc
			},
			wantErr: nil, // freezing a frozen account is idempotent
		},
		{
			name: "freeze closed account",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				acc.Close()
				return acc
			},
			wantErr: ErrAccountClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := tt.setup()
			err := acc.Freeze()
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if acc.Status != StatusFrozen {
				t.Errorf("expected FROZEN, got %s", acc.Status)
			}
		})
	}
}

func TestAccount_Unfreeze_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Account
		wantErr error
	}{
		{
			name: "unfreeze frozen account",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				acc.Freeze()
				return acc
			},
			wantErr: nil,
		},
		{
			name: "unfreeze closed account",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				acc.Close()
				return acc
			},
			wantErr: ErrAccountClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := tt.setup()
			err := acc.Unfreeze()
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if acc.Status != StatusActive {
				t.Errorf("expected ACTIVE, got %s", acc.Status)
			}
		})
	}
}

func TestAccount_Close_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Account
		wantErr error
	}{
		{
			name: "close zero-balance account",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				return acc
			},
			wantErr: nil,
		},
		{
			name: "close non-zero balance",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				acc.Deposit(domain.Money{Amount: 1, Currency: domain.UZS})
				return acc
			},
			wantErr: ErrBalanceNotZero,
		},
		{
			name: "close after deposit and full withdrawal",
			setup: func() *Account {
				acc, _ := NewAccount("user-1", domain.UZS)
				acc.Deposit(domain.Money{Amount: 500000, Currency: domain.UZS})
				acc.Withdraw(domain.Money{Amount: 500000, Currency: domain.UZS})
				return acc
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acc := tt.setup()
			err := acc.Close()
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if acc.Status != StatusClosed {
				t.Errorf("expected CLOSED, got %s", acc.Status)
			}
		})
	}
}

func TestAccount_EventVersioning(t *testing.T) {
	acc, _ := NewAccount("user-1", domain.UZS)

	// After creation: version 1
	if acc.Version != 1 {
		t.Errorf("expected version 1, got %d", acc.Version)
	}

	acc.Deposit(domain.Money{Amount: 100000, Currency: domain.UZS})
	if acc.Version != 2 {
		t.Errorf("expected version 2, got %d", acc.Version)
	}

	acc.Withdraw(domain.Money{Amount: 50000, Currency: domain.UZS})
	if acc.Version != 3 {
		t.Errorf("expected version 3, got %d", acc.Version)
	}

	acc.Freeze()
	if acc.Version != 4 {
		t.Errorf("expected version 4, got %d", acc.Version)
	}

	acc.Unfreeze()
	if acc.Version != 5 {
		t.Errorf("expected version 5, got %d", acc.Version)
	}

	events := acc.UncommittedEvents()
	if len(events) != 5 {
		t.Errorf("expected 5 uncommitted events, got %d", len(events))
	}

	// Verify event types in order
	expectedTypes := []EventType{
		EventAccountOpened,
		EventCredited,
		EventDebited,
		EventFrozen,
		EventUnfrozen,
	}
	for i, et := range expectedTypes {
		if events[i].Type != et {
			t.Errorf("event[%d]: expected type %s, got %s", i, et, events[i].Type)
		}
		if events[i].Version != i+1 {
			t.Errorf("event[%d]: expected version %d, got %d", i, i+1, events[i].Version)
		}
	}
}

func TestAccount_SnapshotRoundTrip(t *testing.T) {
	// Build an account with several operations
	acc, _ := NewAccount("user-1", domain.UZS)
	acc.Deposit(domain.Money{Amount: 1000000, Currency: domain.UZS})
	acc.Withdraw(domain.Money{Amount: 200000, Currency: domain.UZS})

	// Take snapshot
	snap := acc.ToSnapshotState()

	// Restore from snapshot with no additional events
	restored := LoadFromSnapshot(snap, acc.Version, nil)

	if restored.UserID != acc.UserID {
		t.Errorf("UserID: expected %s, got %s", acc.UserID, restored.UserID)
	}
	if restored.AccountNumber != acc.AccountNumber {
		t.Errorf("AccountNumber: expected %s, got %s", acc.AccountNumber, restored.AccountNumber)
	}
	if restored.Balance.Amount != acc.Balance.Amount {
		t.Errorf("Balance: expected %d, got %d", acc.Balance.Amount, restored.Balance.Amount)
	}
	if restored.Status != acc.Status {
		t.Errorf("Status: expected %s, got %s", acc.Status, restored.Status)
	}
}
