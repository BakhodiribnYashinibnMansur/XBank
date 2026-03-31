package card

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrCardNotFound        = apperror.ErrCardNotFound
	ErrCardBlocked         = apperror.ErrCardBlocked
	ErrCardExpired         = apperror.ErrCardExpired
	ErrInvalidPIN          = apperror.ErrInvalidPIN
	ErrPINAttemptsExceeded = apperror.ErrPINAttemptsExceeded
)

type Type string

const (
	TypeDebit   Type = "DEBIT"
	TypeVirtual Type = "VIRTUAL"
)

type Status string

const (
	StatusInactive  Status = "INACTIVE"
	StatusActive    Status = "ACTIVE"
	StatusBlocked   Status = "BLOCKED"
	StatusExpired   Status = "EXPIRED"
	StatusCancelled Status = "CANCELLED"
)

// Card - bank card linked to an account
type Card struct {
	ID            string
	AccountID     string
	PAN           string // full card number (stored encrypted in real DB)
	MaskedPAN     string // **** **** **** 1234
	PINHash       string // bcrypt hashed PIN
	ExpiryMonth   int
	ExpiryYear    int
	CardType      Type
	Status        Status
	PINAttempts   int // wrong PIN attempt counter
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// NewCard - create a new card (INACTIVE by default, needs activation)
func NewCard(accountID string, cardType Type) (*Card, error) {
	if accountID == "" {
		return nil, apperror.ErrMissingField.WithMessage("account_id cannot be empty")
	}

	pan, err := GenerateCardNumber()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &Card{
		AccountID:   accountID,
		PAN:         pan,
		MaskedPAN:   MaskPAN(pan),
		CardType:    cardType,
		Status:      StatusInactive,
		ExpiryMonth: int(now.Month()),
		ExpiryYear:  now.Year() + 3, // 3 years validity
		PINAttempts: 0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Activate - activate the card by setting a PIN
func (c *Card) Activate(pin string) error {
	if c.Status != StatusInactive {
		return apperror.ErrValidation.WithMessage("Only inactive cards can be activated")
	}

	if len(pin) != 4 {
		return ErrInvalidPIN
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	c.PINHash = string(hash)
	c.Status = StatusActive
	c.UpdatedAt = time.Now()
	return nil
}

// VerifyPIN - check the PIN (with brute-force protection)
func (c *Card) VerifyPIN(pin string) error {
	if c.Status == StatusBlocked {
		return ErrCardBlocked
	}
	if c.PINAttempts >= 3 {
		c.Status = StatusBlocked
		return ErrPINAttemptsExceeded
	}

	if err := bcrypt.CompareHashAndPassword([]byte(c.PINHash), []byte(pin)); err != nil {
		c.PINAttempts++
		if c.PINAttempts >= 3 {
			c.Status = StatusBlocked
			return ErrPINAttemptsExceeded
		}
		return ErrInvalidPIN
	}

	// Successful PIN entry resets the counter
	c.PINAttempts = 0
	return nil
}

// ChangePin - change the card PIN (must verify old PIN first)
func (c *Card) ChangePIN(oldPIN, newPIN string) error {
	if err := c.checkUsable(); err != nil {
		return err
	}

	if err := c.VerifyPIN(oldPIN); err != nil {
		return err
	}

	if len(newPIN) != 4 {
		return ErrInvalidPIN
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPIN), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	c.PINHash = string(hash)
	c.UpdatedAt = time.Now()
	return nil
}

// Block - block the card
func (c *Card) Block() error {
	if c.Status == StatusCancelled {
		return apperror.ErrValidation.WithMessage("Cannot block a cancelled card")
	}
	c.Status = StatusBlocked
	c.UpdatedAt = time.Now()
	return nil
}

// Unblock - unblock the card and reset PIN attempts
func (c *Card) Unblock() error {
	if c.Status != StatusBlocked {
		return apperror.ErrValidation.WithMessage("Card is not blocked")
	}
	c.Status = StatusActive
	c.PINAttempts = 0
	c.UpdatedAt = time.Now()
	return nil
}

// Cancel - permanently cancel the card
func (c *Card) Cancel() error {
	c.Status = StatusCancelled
	c.UpdatedAt = time.Now()
	return nil
}

// IsExpired - check if the card has expired
func (c *Card) IsExpired() bool {
	now := time.Now()
	return now.Year() > c.ExpiryYear ||
		(now.Year() == c.ExpiryYear && int(now.Month()) > c.ExpiryMonth)
}

func (c *Card) checkUsable() error {
	switch c.Status {
	case StatusInactive:
		return apperror.ErrValidation.WithMessage("Card is not activated")
	case StatusBlocked:
		return ErrCardBlocked
	case StatusExpired:
		return ErrCardExpired
	case StatusCancelled:
		return apperror.ErrValidation.WithMessage("Card is cancelled")
	}
	if c.IsExpired() {
		c.Status = StatusExpired
		return ErrCardExpired
	}
	return nil
}

// Repository - interface for card persistence
type Repository interface {
	Create(ctx context.Context, card *Card) error
	GetByID(ctx context.Context, id string) (*Card, error)
	ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*Card, error)
	CountByAccountID(ctx context.Context, accountID string) (int64, error)
	Update(ctx context.Context, card *Card) error
}
