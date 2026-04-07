package domain

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"
)

// HoldStatus - authorization hold lifecycle
type HoldStatus string

const (
	HoldStatusHeld     HoldStatus = "HELD"     // funds reserved
	HoldStatusCaptured HoldStatus = "CAPTURED"  // merchant settled
	HoldStatusReleased HoldStatus = "RELEASED"  // hold cancelled
	HoldStatusExpired  HoldStatus = "EXPIRED"   // TTL expired
)

const DefaultHoldTTL = 7 * 24 * time.Hour // 7 days

// Hold - an authorization hold on a card.
//
// Flow:
//  1. Hold    — reserve funds (available_balance decreases, balance unchanged)
//  2. Capture — settle the hold (balance decreases, hold removed)
//  3. Release — cancel the hold (available_balance restored)
//
// Holds expire after 7 days if neither captured nor released.
type Hold struct {
	ID         string
	CardID     string
	AccountID  string
	Merchant   string
	Amount     int64
	Currency   string
	Status     HoldStatus
	HeldAt     time.Time
	ExpiresAt  time.Time
	CapturedAt *time.Time
	ReleasedAt *time.Time
}

// NewHold - create a new authorization hold
func NewHold(cardID, accountID, merchant string, amount int64, currency string) (*Hold, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("hold amount must be positive")
	}

	id, _ := generateHoldID()
	now := time.Now()

	return &Hold{
		ID:        id,
		CardID:    cardID,
		AccountID: accountID,
		Merchant:  merchant,
		Amount:    amount,
		Currency:  currency,
		Status:    HoldStatusHeld,
		HeldAt:    now,
		ExpiresAt: now.Add(DefaultHoldTTL),
	}, nil
}

// Capture - settle the authorization hold
func (h *Hold) Capture() error {
	if h.Status != HoldStatusHeld {
		return fmt.Errorf("can only capture a held authorization (current: %s)", h.Status)
	}
	if time.Now().After(h.ExpiresAt) {
		h.Status = HoldStatusExpired
		return fmt.Errorf("hold has expired")
	}
	h.Status = HoldStatusCaptured
	now := time.Now()
	h.CapturedAt = &now
	return nil
}

// Release - cancel the authorization hold (refund reserved funds)
func (h *Hold) Release() error {
	if h.Status != HoldStatusHeld {
		return fmt.Errorf("can only release a held authorization (current: %s)", h.Status)
	}
	h.Status = HoldStatusReleased
	now := time.Now()
	h.ReleasedAt = &now
	return nil
}

// IsExpired - check if the hold has expired
func (h *Hold) IsExpired() bool {
	return h.Status == HoldStatusHeld && time.Now().After(h.ExpiresAt)
}

// HoldRepository - persistence for authorization holds
type HoldRepository interface {
	Create(ctx context.Context, hold *Hold) error
	GetByID(ctx context.Context, id string) (*Hold, error)
	ListByCardID(ctx context.Context, cardID string) ([]*Hold, error)
	ListActiveByAccountID(ctx context.Context, accountID string) ([]*Hold, error)
	Update(ctx context.Context, hold *Hold) error
	FetchExpired(ctx context.Context, limit int) ([]*Hold, error)
}

func generateHoldID() (string, error) {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}
