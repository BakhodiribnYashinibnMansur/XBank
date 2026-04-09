package domain

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

var (
	ErrCurrencyNotFound = domain.NewDomainError("CURRENCY_NOT_FOUND", "currency not found")
	ErrCurrencyExists   = domain.NewDomainError("CURRENCY_EXISTS", "currency already exists")
	ErrMissingCode      = domain.NewDomainError("MISSING_FIELD", "currency code cannot be empty")
	ErrInvalidCode      = domain.NewDomainError("INVALID_CODE", "currency code must be 3 uppercase letters")
)

// Status represents the currency availability status.
type Status string

const (
	StatusActive   Status = "ACTIVE"
	StatusInactive Status = "INACTIVE"
)

// Currency represents a currency in the system catalog.
type Currency struct {
	ID            string
	Code          string // ISO 4217 code, e.g. "USD", "UZS"
	Name          string // e.g. "US Dollar"
	Symbol        string // e.g. "$", "so'm"
	DecimalPlaces int    // e.g. 2 for USD, 0 for UZS
	Status        Status
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewCurrency creates a new currency entity with validation.
func NewCurrency(code, name, symbol string, decimalPlaces int) (*Currency, error) {
	if code == "" {
		return nil, ErrMissingCode
	}
	if len(code) != 3 {
		return nil, ErrInvalidCode
	}
	return &Currency{
		Code:          code,
		Name:          name,
		Symbol:        symbol,
		DecimalPlaces: decimalPlaces,
		Status:        StatusActive,
	}, nil
}

// Activate marks the currency as active.
func (c *Currency) Activate() {
	c.Status = StatusActive
}

// Deactivate marks the currency as inactive.
func (c *Currency) Deactivate() {
	c.Status = StatusInactive
}

// Repository defines the persistence interface for currencies.
type Repository interface {
	Create(ctx context.Context, currency *Currency) error
	GetByID(ctx context.Context, id string) (*Currency, error)
	GetByCode(ctx context.Context, code string) (*Currency, error)
	ListAll(ctx context.Context) ([]*Currency, error)
	Update(ctx context.Context, currency *Currency) error
}
