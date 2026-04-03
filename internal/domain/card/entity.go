package card

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrCardNotFound        = shared.NewDomainError("CARD_NOT_FOUND", "card not found")
	ErrCardBlocked         = shared.NewDomainError("CARD_BLOCKED", "card is blocked")
	ErrCardExpired         = shared.NewDomainError("CARD_EXPIRED", "card has expired")
	ErrInvalidPIN          = shared.NewDomainError("INVALID_PIN", "invalid PIN")
	ErrPINAttemptsExceeded = shared.NewDomainError("PIN_ATTEMPTS_EXCEEDED", "card locked: too many wrong PIN attempts")
	ErrCardValidation      = shared.NewDomainError("CARD_VALIDATION", "card validation error")
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

	// 3D Secure / EMV
	ThreeDSEnrolled bool   // enrolled in 3D Secure
	ThreeDSVersion  string // "2.1", "2.2" etc.
	EMVAID          string // Application Identifier (e.g. "A0000000041010" for Mastercard)

	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewCard - create a new card (INACTIVE by default, needs activation)
func NewCard(accountID string, cardType Type) (*Card, error) {
	if accountID == "" {
		return nil, shared.NewDomainError("MISSING_FIELD", "account_id cannot be empty")
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
		return shared.NewDomainError("CARD_VALIDATION", "only inactive cards can be activated")
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
		return shared.NewDomainError("CARD_VALIDATION", "cannot block a cancelled card")
	}
	c.Status = StatusBlocked
	c.UpdatedAt = time.Now()
	return nil
}

// Unblock - unblock the card and reset PIN attempts
func (c *Card) Unblock() error {
	if c.Status != StatusBlocked {
		return shared.NewDomainError("CARD_VALIDATION", "card is not blocked")
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
		return shared.NewDomainError("CARD_VALIDATION", "card is not activated")
	case StatusBlocked:
		return ErrCardBlocked
	case StatusExpired:
		return ErrCardExpired
	case StatusCancelled:
		return shared.NewDomainError("CARD_VALIDATION", "card is cancelled")
	}
	if c.IsExpired() {
		c.Status = StatusExpired
		return ErrCardExpired
	}
	return nil
}

// Enroll3DS - enroll the card in 3D Secure
func (c *Card) Enroll3DS(version string) error {
	if err := c.checkUsable(); err != nil {
		return err
	}
	c.ThreeDSEnrolled = true
	c.ThreeDSVersion = version
	c.UpdatedAt = time.Now()
	return nil
}

// SetEMVAID - set the EMV Application Identifier
func (c *Card) SetEMVAID(aid string) error {
	if err := c.checkUsable(); err != nil {
		return err
	}
	c.EMVAID = aid
	c.UpdatedAt = time.Now()
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
