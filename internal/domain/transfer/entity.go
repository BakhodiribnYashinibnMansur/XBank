package transfer

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
)

var (
	ErrTransferNotFound = apperror.ErrTransferNotFound
	ErrSameAccount      = apperror.ErrSameAccount
	ErrInvalidAmount    = apperror.ErrInvalidAmount
	ErrTransferFailed   = apperror.ErrTransferFailed
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusCompleted Status = "COMPLETED"
	StatusFailed    Status = "FAILED"
)

// Transfer - event-sourced money transfer aggregate
type Transfer struct {
	ID            string
	FromAccountID string
	ToAccountID   string
	Amount        shared.Money
	Status        Status
	Description   string
	FailureReason string
	Version       int // current version = number of events applied
	CreatedAt     time.Time

	uncommittedEvents []Event // events produced but not yet persisted
}

// --- Factory ---

// NewTransfer - create a new transfer (raises TransferCreated event)
func NewTransfer(fromAccountID, toAccountID string, amount shared.Money, description string) (*Transfer, error) {
	if fromAccountID == toAccountID {
		return nil, ErrSameAccount
	}
	if amount.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	t := &Transfer{}
	t.ID = generateUUID()
	t.raise(EventTransferCreated, TransferCreatedData{
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        amount.Amount,
		Currency:      string(amount.Currency),
		Description:   description,
	})
	return t, nil
}

// LoadFromHistory - rebuild transfer state from events
func LoadFromHistory(events []Event) *Transfer {
	t := &Transfer{}
	for _, e := range events {
		t.Apply(e)
	}
	return t
}

// --- Commands (validate + raise event) ---

// Complete - mark the transfer as successfully completed
func (t *Transfer) Complete() {
	t.raise(EventTransferCompleted, TransferCompletedData{})
}

// Fail - mark the transfer as failed
func (t *Transfer) Fail(reason string) {
	t.raise(EventTransferFailed, TransferFailedData{Reason: reason})
}

// --- Event application (pure state transition, no validation) ---

func (t *Transfer) Apply(e Event) {
	switch data := e.Data.(type) {
	case TransferCreatedData:
		t.ID = e.AggregateID
		t.FromAccountID = data.FromAccountID
		t.ToAccountID = data.ToAccountID
		t.Amount = shared.Money{Amount: data.Amount, Currency: shared.Currency(data.Currency)}
		t.Description = data.Description
		t.Status = StatusPending
		t.CreatedAt = e.OccurredAt
	case TransferCompletedData:
		t.Status = StatusCompleted
	case TransferFailedData:
		t.Status = StatusFailed
		t.FailureReason = data.Reason
	}
	t.Version = e.Version
}

// --- Uncommitted events ---

func (t *Transfer) UncommittedEvents() []Event {
	return t.uncommittedEvents
}

func (t *Transfer) ClearUncommittedEvents() {
	t.uncommittedEvents = nil
}

// --- Internal helpers ---

func (t *Transfer) raise(eventType EventType, data EventData) {
	e := Event{
		AggregateID: t.ID,
		Type:        eventType,
		Data:        data,
		Version:     t.Version + 1,
		OccurredAt:  time.Now(),
	}
	t.Apply(e)
	t.uncommittedEvents = append(t.uncommittedEvents, e)
}

// Repository - read projection interface (existing transfers table)
type Repository interface {
	Create(ctx context.Context, transfer *Transfer) error
	GetByID(ctx context.Context, id string) (*Transfer, error)
	ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*Transfer, error)
	CountByAccountID(ctx context.Context, accountID string) (int64, error)
	Update(ctx context.Context, transfer *Transfer) error
}

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
