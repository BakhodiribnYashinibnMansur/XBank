package domain

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"
)

// Method - how the user proves identity for step-up auth
type Method string

const (
	MethodPassword Method = "PASSWORD" // re-enter password
	MethodTOTP     Method = "TOTP"     // TOTP code (future)
)

// Status - challenge lifecycle
type Status string

const (
	StatusPending  Status = "PENDING"
	StatusVerified Status = "VERIFIED"
	StatusExpired  Status = "EXPIRED"
	StatusFailed   Status = "FAILED"
)

// Challenge - step-up authentication for sensitive operations.
//
// Flow:
//  1. User requests a sensitive action (e.g. large transfer)
//  2. Server creates a Challenge (PENDING)
//  3. User verifies identity (password/TOTP)
//  4. Server marks Challenge as VERIFIED and returns a token
//  5. User includes X-Challenge-Token in the sensitive request
//  6. Middleware validates the token and allows the request
type Challenge struct {
	ID        string
	UserID    string
	Method    Method
	Status    Status
	Token     string    // opaque token returned after verification
	Action    string    // what action this challenge authorizes (e.g. "transfer")
	Metadata  string    // action details (e.g. transfer_id, amount) — for audit
	ExpiresAt time.Time // challenge must be verified before this
	CreatedAt time.Time
	VerifiedAt *time.Time
}

const (
	DefaultTTL      = 5 * time.Minute  // challenge expires in 5 min
	TokenTTL        = 10 * time.Minute // verified token valid for 10 min
	MaxAttempts     = 3                // max wrong verifications
)

// NewChallenge - create a new pending challenge
func NewChallenge(userID string, method Method, action, metadata string) (*Challenge, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	id, err := generateID()
	if err != nil {
		return nil, err
	}

	return &Challenge{
		ID:        id,
		UserID:    userID,
		Method:    method,
		Status:    StatusPending,
		Action:    action,
		Metadata:  metadata,
		ExpiresAt: time.Now().Add(DefaultTTL),
		CreatedAt: time.Now(),
	}, nil
}

// Verify - mark as verified and generate token
func (c *Challenge) Verify() error {
	if c.Status != StatusPending {
		return fmt.Errorf("challenge is not pending (status: %s)", c.Status)
	}
	if time.Now().After(c.ExpiresAt) {
		c.Status = StatusExpired
		return fmt.Errorf("challenge has expired")
	}

	token, err := generateToken()
	if err != nil {
		return err
	}

	c.Status = StatusVerified
	c.Token = token
	now := time.Now()
	c.VerifiedAt = &now
	c.ExpiresAt = now.Add(TokenTTL) // extend expiry for token usage
	return nil
}

// Fail - mark as failed (wrong password/code)
func (c *Challenge) Fail() {
	c.Status = StatusFailed
}

// IsTokenValid - check if the challenge token is still usable
func (c *Challenge) IsTokenValid(token string) bool {
	return c.Status == StatusVerified &&
		c.Token == token &&
		time.Now().Before(c.ExpiresAt)
}

// Repository - persistence interface
type Repository interface {
	Create(ctx context.Context, challenge *Challenge) error
	GetByID(ctx context.Context, id string) (*Challenge, error)
	GetByToken(ctx context.Context, token string) (*Challenge, error)
	Update(ctx context.Context, challenge *Challenge) error
	CountPendingByUser(ctx context.Context, userID string) (int, error)
}

func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", b), nil
}
