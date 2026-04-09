package apikey

import (
	"context"
	"time"
)

// Key represents an API key record.
type Key struct {
	ID        string
	Prefix    string // first 8 chars after the fixed prefix, for lookup
	Hash      string // SHA-256 of the full raw key
	UserID    string
	Scopes    []string
	ExpiresAt *time.Time
	CreatedAt time.Time
	Revoked   bool
}

// IsExpired reports whether the key has passed its expiration time.
func (k *Key) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}

// Store defines the persistence interface for API keys.
// Implementation lives in the infrastructure/postgres layer.
type Store interface {
	Save(ctx context.Context, key *Key) error
	GetByPrefix(ctx context.Context, prefix string) (*Key, error)
	Revoke(ctx context.Context, id string) error
}
