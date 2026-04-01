package shared

import (
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
)

var (
	ErrNegativeAmount    = apperror.ErrInvalidAmount
	ErrCurrencyMismatch  = apperror.ErrCurrencyMismatch
	ErrInsufficientFunds = apperror.ErrInsufficientFunds
)

// Currency - currency code (ISO 4217)
type Currency string

const (
	UZS Currency = "UZS" // Uzbek som (exponent: 2, i.e. tiyin)
	USD Currency = "USD" // US dollar (exponent: 2, i.e. cent)
	EUR Currency = "EUR" // Euro (exponent: 2, i.e. cent)
)

// Money - monetary value (Value Object)
// Amount is ALWAYS stored in minor units (tiyin/cent)
// Example: 15000.50 UZS = 1500050 tiyin
type Money struct {
	Amount   int64    // in tiyin/cent
	Currency Currency
}

// NewMoney - create a new Money value
func NewMoney(amount int64, currency Currency) (Money, error) {
	if amount < 0 {
		return Money{}, ErrNegativeAmount
	}
	return Money{Amount: amount, Currency: currency}, nil
}

// Add - add two monetary values (currencies must match)
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{Amount: m.Amount + other.Amount, Currency: m.Currency}, nil
}

// Subtract - subtract (sufficient funds required)
func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	if m.Amount < other.Amount {
		return Money{}, ErrInsufficientFunds
	}
	return Money{Amount: m.Amount - other.Amount, Currency: m.Currency}, nil
}

// IsZero - check if amount is zero
func (m Money) IsZero() bool {
	return m.Amount == 0
}

// String - pretty format: "15000.50 UZS"
func (m Money) String() string {
	major := m.Amount / 100
	minor := m.Amount % 100
	return fmt.Sprintf("%d.%02d %s", major, minor, m.Currency)
}