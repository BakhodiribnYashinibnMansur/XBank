package auth

import (
	"crypto/ecdsa"
	"fmt"
	"sync"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.uber.org/zap"
)

// KeyRotator - manages JWT key rotation
// New key for signing, old keys kept for verification during grace period
type KeyRotator struct {
	mu          sync.RWMutex
	currentKey  *ecdsa.PrivateKey
	currentPub  *ecdsa.PublicKey
	oldKeys     []*ecdsa.PublicKey // old public keys still valid for verification
	gracePeriod time.Duration
}

func NewKeyRotator(privateKeyPath, publicKeyPath string, gracePeriod time.Duration) (*KeyRotator, error) {
	priv, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, err
	}
	pub, err := loadPublicKey(publicKeyPath)
	if err != nil {
		return nil, err
	}
	return &KeyRotator{
		currentKey:  priv,
		currentPub:  pub,
		gracePeriod: gracePeriod,
	}, nil
}

// Rotate - load new keys, keep old for grace period
func (r *KeyRotator) Rotate(privateKeyPath, publicKeyPath string) error {
	newPriv, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		return fmt.Errorf("rotate: load private key: %w", err)
	}
	newPub, err := loadPublicKey(publicKeyPath)
	if err != nil {
		return fmt.Errorf("rotate: load public key: %w", err)
	}

	r.mu.Lock()
	// Move current public key to old keys
	r.oldKeys = append(r.oldKeys, r.currentPub)
	r.currentKey = newPriv
	r.currentPub = newPub
	r.mu.Unlock()

	logger.Log.Info("JWT key rotated, old key valid for grace period",
		zap.Duration("grace_period", r.gracePeriod),
	)

	// Remove old key after grace period
	go func() {
		time.Sleep(r.gracePeriod)
		r.mu.Lock()
		if len(r.oldKeys) > 0 {
			r.oldKeys = r.oldKeys[1:] // remove oldest
		}
		r.mu.Unlock()
		logger.Log.Info("old JWT key expired and removed")
	}()

	return nil
}

// SigningKey - current private key for signing
func (r *KeyRotator) SigningKey() *ecdsa.PrivateKey {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.currentKey
}

// VerificationKeys - current + old public keys for verification
func (r *KeyRotator) VerificationKeys() []*ecdsa.PublicKey {
	r.mu.RLock()
	defer r.mu.RUnlock()
	keys := make([]*ecdsa.PublicKey, 0, len(r.oldKeys)+1)
	keys = append(keys, r.currentPub)
	keys = append(keys, r.oldKeys...)
	return keys
}
