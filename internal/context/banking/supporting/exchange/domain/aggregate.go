package domain

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
)

var (
	ErrRateNotFound = shared.NewDomainError("RATE_NOT_FOUND", "exchange rate not found")
)

// Rate - exchange rate between two currencies
type Rate struct {
	ID           string
	FromCurrency string // e.g. "USD"
	ToCurrency   string // e.g. "UZS"
	BuyRate      int64  // bank buys at this rate (minor units, e.g. 1265050 = 12650.50)
	SellRate     int64  // bank sells at this rate
	UpdatedAt    time.Time
}

// Convert - convert amount from one currency to another (using sell rate)
// amount in minor units, returns minor units
func (r *Rate) Convert(amount int64) int64 {
	// SellRate is per 1 unit of FromCurrency in minor units of ToCurrency
	// e.g. 1 USD = 12650.50 UZS → SellRate = 1265050
	return amount * r.SellRate / 100
}

type Repository interface {
	Upsert(ctx context.Context, rate *Rate) error
	GetRate(ctx context.Context, from, to string) (*Rate, error)
	ListAll(ctx context.Context) ([]*Rate, error)
}
