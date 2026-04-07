package domain

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
)

var (
	ErrAccountNotFound = shared.NewDomainError("ACCOUNT_NOT_FOUND", "account not found")
	ErrAccountFrozen   = shared.NewDomainError("ACCOUNT_FROZEN", "account is frozen")
	ErrAccountClosed   = shared.NewDomainError("ACCOUNT_CLOSED", "account is closed")
	ErrBalanceNotZero  = shared.NewDomainError("BALANCE_NOT_ZERO", "account balance must be 0 to close")
	ErrMissingUserID   = shared.NewDomainError("MISSING_FIELD", "user_id cannot be empty")
)

type Status string

const (
	StatusActive Status = "ACTIVE"
	StatusFrozen Status = "FROZEN"
	StatusClosed Status = "CLOSED"
)

// Account - event-sourced bank account aggregate
type Account struct {
	ID            string
	UserID        string
	AccountNumber string
	Balance       shared.Money
	Status        Status
	Version       int // current version = number of events applied
	CreatedAt     time.Time
	UpdatedAt     time.Time

	uncommittedEvents []Event // events produced but not yet persisted
}

// --- Factory ---

// NewAccount - create a new account (raises AccountOpened event)
func NewAccount(userID string, currency shared.Currency) (*Account, error) {
	if userID == "" {
		return nil, ErrMissingUserID
	}

	accountNumber, err := generateAccountNumber()
	if err != nil {
		return nil, err
	}

	a := &Account{}
	a.ID = generateUUID()
	a.raise(EventAccountOpened, AccountOpenedData{
		UserID:        userID,
		AccountNumber: accountNumber,
		Currency:      string(currency),
	})
	return a, nil
}

// LoadFromHistory - rebuild account state from events
func LoadFromHistory(events []Event) *Account {
	a := &Account{}
	for _, e := range events {
		a.Apply(e)
	}
	return a
}

// LoadFromSnapshot - rebuild account from snapshot + remaining events
func LoadFromSnapshot(snap SnapshotState, version int, events []Event) *Account {
	a := &Account{
		UserID:        snap.UserID,
		AccountNumber: snap.AccountNumber,
		Balance:       shared.Money{Amount: snap.Balance, Currency: shared.Currency(snap.Currency)},
		Status:        Status(snap.Status),
		Version:       version,
	}
	for _, e := range events {
		a.Apply(e)
	}
	return a
}

// --- Commands (validate + raise event) ---

func (a *Account) Deposit(amount shared.Money) error {
	if err := a.checkActive(); err != nil {
		return err
	}
	if a.Balance.Currency != amount.Currency {
		return shared.ErrCurrencyMismatch
	}
	a.raise(EventCredited, CreditedData{
		Amount:   amount.Amount,
		Currency: string(amount.Currency),
	})
	return nil
}

func (a *Account) Withdraw(amount shared.Money) error {
	if err := a.checkActive(); err != nil {
		return err
	}
	if a.Balance.Currency != amount.Currency {
		return shared.ErrCurrencyMismatch
	}
	if a.Balance.Amount < amount.Amount {
		return shared.ErrInsufficientFunds
	}
	a.raise(EventDebited, DebitedData{
		Amount:   amount.Amount,
		Currency: string(amount.Currency),
	})
	return nil
}

func (a *Account) Freeze() error {
	if a.Status == StatusClosed {
		return ErrAccountClosed
	}
	a.raise(EventFrozen, FrozenData{})
	return nil
}

func (a *Account) Unfreeze() error {
	if a.Status == StatusClosed {
		return ErrAccountClosed
	}
	a.raise(EventUnfrozen, UnfrozenData{})
	return nil
}

func (a *Account) Close() error {
	if !a.Balance.IsZero() {
		return ErrBalanceNotZero
	}
	a.raise(EventClosed, ClosedData{})
	return nil
}

// --- Event application (pure state transition, no validation) ---

func (a *Account) Apply(e Event) {
	switch data := e.Data.(type) {
	case AccountOpenedData:
		a.ID = e.AggregateID
		a.UserID = data.UserID
		a.AccountNumber = data.AccountNumber
		a.Balance = shared.Money{Amount: 0, Currency: shared.Currency(data.Currency)}
		a.Status = StatusActive
		a.CreatedAt = e.OccurredAt
	case CreditedData:
		a.Balance.Amount += data.Amount
	case DebitedData:
		a.Balance.Amount -= data.Amount
	case FrozenData:
		a.Status = StatusFrozen
	case UnfrozenData:
		a.Status = StatusActive
	case ClosedData:
		a.Status = StatusClosed
	}
	a.Version = e.Version
	a.UpdatedAt = e.OccurredAt
}

// --- Uncommitted events ---

func (a *Account) UncommittedEvents() []Event {
	return a.uncommittedEvents
}

func (a *Account) ClearUncommittedEvents() {
	a.uncommittedEvents = nil
}

// ToSnapshotState - convert current state for snapshot persistence
func (a *Account) ToSnapshotState() SnapshotState {
	return SnapshotState{
		UserID:        a.UserID,
		AccountNumber: a.AccountNumber,
		Balance:       a.Balance.Amount,
		Currency:      string(a.Balance.Currency),
		Status:        string(a.Status),
	}
}

// --- Internal helpers ---

func (a *Account) raise(eventType EventType, data EventData) {
	e := Event{
		AggregateID: a.ID,
		Type:        eventType,
		Data:        data,
		Version:     a.Version + 1,
		OccurredAt:  time.Now(),
	}
	a.Apply(e)
	a.uncommittedEvents = append(a.uncommittedEvents, e)
}

func (a *Account) checkActive() error {
	switch a.Status {
	case StatusFrozen:
		return ErrAccountFrozen
	case StatusClosed:
		return ErrAccountClosed
	}
	return nil
}

func generateAccountNumber() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%016x", bytes)[:16], nil
}

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// Repository - read projection interface (existing accounts table)
type Repository interface {
	Create(ctx context.Context, account *Account) error
	GetByID(ctx context.Context, id string) (*Account, error)
	GetByIDForUpdate(ctx context.Context, id string) (*Account, error)
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*Account, error)
	CountByUserID(ctx context.Context, userID string) (int64, error)
	Update(ctx context.Context, account *Account) error
}
