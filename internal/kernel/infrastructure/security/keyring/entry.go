package keyring

import "time"

// KeyType identifies the cryptographic algorithm family.
type KeyType string

const (
	KeyTypeECDSA KeyType = "ecdsa"
	KeyTypeHMAC  KeyType = "hmac"
	KeyTypeAES   KeyType = "aes"
)

// KeyEntry represents a single versioned key in the keyring.
type KeyEntry struct {
	ID        string
	Type      KeyType
	Key       any // *ecdsa.PrivateKey, []byte, etc.
	PublicKey any // *ecdsa.PublicKey or nil for symmetric keys
	Version   int
	Active    bool
	CreatedAt time.Time
	ExpiresAt time.Time
}

// IsExpired reports whether the key has passed its expiration time.
func (e *KeyEntry) IsExpired() bool {
	if e.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(e.ExpiresAt)
}
