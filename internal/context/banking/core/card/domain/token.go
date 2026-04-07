package domain

import (
	"context"
	"time"
)

// Token - an opaque token that represents a card PAN.
// Used by merchants/payment systems instead of real PAN.
type Token struct {
	Token        string
	CardID       string
	PANEncrypted string // AES-256-GCM encrypted PAN
	LastFour     string
	ExpiresAt    time.Time
	CreatedAt    time.Time
	IsActive     bool
}

// TokenRepository - persistence for card tokens
type TokenRepository interface {
	Create(ctx context.Context, token *Token) error
	GetByToken(ctx context.Context, token string) (*Token, error)
	ListByCardID(ctx context.Context, cardID string) ([]*Token, error)
	Deactivate(ctx context.Context, token string) error
}
