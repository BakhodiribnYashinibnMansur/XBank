package session

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
)

var (
	ErrSessionNotFound = shared.NewDomainError("SESSION_NOT_FOUND", "session not found")
	ErrSessionExpired  = shared.NewDomainError("SESSION_EXPIRED", "session has expired")
	ErrInvalidToken    = shared.NewDomainError("INVALID_TOKEN", "invalid token")
)

// Session - user session
// A new session is created on each login
type Session struct {
	ID           string
	UserID       string
	RefreshToken string // hashed (SHA-256)
	UserAgent    string
	IPAddress    string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

// NewSession - create a new session
func NewSession(userID, refreshTokenHash, userAgent, ipAddress string, expiresAt time.Time) (*Session, error) {
	if userID == "" {
		return nil, shared.NewDomainError("MISSING_FIELD", "user_id cannot be empty")
	}
	if refreshTokenHash == "" {
		return nil, ErrInvalidToken
	}

	return &Session{
		UserID:       userID,
		RefreshToken: refreshTokenHash,
		UserAgent:    userAgent,
		IPAddress:    ipAddress,
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
	}, nil
}

// IsExpired - check if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Repository - interface for working with sessions
type Repository interface {
	Create(ctx context.Context, session *Session) error
	GetByRefreshToken(ctx context.Context, refreshTokenHash string) (*Session, error)
	DeleteByID(ctx context.Context, id string) error
	DeleteAllByUserID(ctx context.Context, userID string) error
}
