package device

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"
)

// Fingerprint - a trusted device record
type Fingerprint struct {
	ID         string
	UserID     string
	DeviceHash string // SHA-256 of device ID
	DeviceName string // User-Agent or device label
	IPAddress  string
	Trusted    bool
	LastUsedAt time.Time
	CreatedAt  time.Time
}

// HashDevice - SHA-256 hash of raw device ID
func HashDevice(deviceID string) string {
	h := sha256.Sum256([]byte(deviceID))
	return fmt.Sprintf("%x", h)
}

type Repository interface {
	Upsert(ctx context.Context, fp *Fingerprint) error
	GetByUserAndDevice(ctx context.Context, userID, deviceHash string) (*Fingerprint, error)
	ListByUserID(ctx context.Context, userID string) ([]*Fingerprint, error)
	Delete(ctx context.Context, id string) error
}
