package keyring

import (
	"fmt"
	"sync"
	"time"
)

// Keyring manages a set of versioned cryptographic keys with thread-safe access.
type Keyring struct {
	mu   sync.RWMutex
	keys map[string]*KeyEntry // keyed by ID
}

// New creates an empty Keyring.
func New() *Keyring {
	return &Keyring{keys: make(map[string]*KeyEntry)}
}

// Add inserts a key entry into the keyring.
func (kr *Keyring) Add(entry *KeyEntry) error {
	if entry.ID == "" {
		return fmt.Errorf("keyring: key ID cannot be empty")
	}

	kr.mu.Lock()
	defer kr.mu.Unlock()

	if _, exists := kr.keys[entry.ID]; exists {
		return fmt.Errorf("keyring: key %q already exists", entry.ID)
	}

	kr.keys[entry.ID] = entry
	return nil
}

// Get retrieves a key by its ID.
func (kr *Keyring) Get(id string) (*KeyEntry, error) {
	kr.mu.RLock()
	defer kr.mu.RUnlock()

	entry, ok := kr.keys[id]
	if !ok {
		return nil, fmt.Errorf("keyring: key %q not found", id)
	}
	return entry, nil
}

// Active returns the currently active key of the given type.
// Returns error if no active key is found.
func (kr *Keyring) Active(keyType KeyType) (*KeyEntry, error) {
	kr.mu.RLock()
	defer kr.mu.RUnlock()

	var best *KeyEntry
	for _, entry := range kr.keys {
		if entry.Type == keyType && entry.Active && !entry.IsExpired() {
			if best == nil || entry.Version > best.Version {
				best = entry
			}
		}
	}

	if best == nil {
		return nil, fmt.Errorf("keyring: no active key of type %q", keyType)
	}
	return best, nil
}

// Rotate deactivates all keys of the given type and activates the new entry.
// During the grace period, the old keys remain in the keyring for verification
// but are marked inactive (no new signatures).
func (kr *Keyring) Rotate(keyType KeyType, newEntry *KeyEntry, gracePeriod time.Duration) error {
	if newEntry.ID == "" {
		return fmt.Errorf("keyring: new key ID cannot be empty")
	}

	kr.mu.Lock()
	defer kr.mu.Unlock()

	// Deactivate old keys of the same type
	for _, entry := range kr.keys {
		if entry.Type == keyType && entry.Active {
			entry.Active = false
			if gracePeriod > 0 && entry.ExpiresAt.IsZero() {
				entry.ExpiresAt = time.Now().Add(gracePeriod)
			}
		}
	}

	newEntry.Active = true
	kr.keys[newEntry.ID] = newEntry
	return nil
}

// Remove deletes a key by ID.
func (kr *Keyring) Remove(id string) {
	kr.mu.Lock()
	defer kr.mu.Unlock()
	delete(kr.keys, id)
}

// AllPublic returns all non-expired key entries of the given type that have a public key.
// Useful for constructing JWKS endpoints.
func (kr *Keyring) AllPublic(keyType KeyType) []*KeyEntry {
	kr.mu.RLock()
	defer kr.mu.RUnlock()

	var result []*KeyEntry
	for _, entry := range kr.keys {
		if entry.Type == keyType && entry.PublicKey != nil && !entry.IsExpired() {
			result = append(result, entry)
		}
	}
	return result
}
