package crypto

import (
	"crypto/rand"
	"fmt"
)

// Tokenizer - generates and manages PAN tokens.
// Real PAN is encrypted with AES-256-GCM and stored in the vault.
// A random, non-reversible token is returned to the caller.
//
// Token format: tok_ + 20 hex chars (e.g. tok_a1b2c3d4e5f6a7b8c9d0)
type Tokenizer struct {
	encryptor *AESEncryptor
}

// NewTokenizer creates a tokenizer backed by AES encryption
func NewTokenizer(encryptor *AESEncryptor) *Tokenizer {
	return &Tokenizer{encryptor: encryptor}
}

// Tokenize - encrypt PAN and generate a random token.
// Returns (token, encryptedPAN, lastFour, error)
func (t *Tokenizer) Tokenize(pan string) (token, encryptedPAN, lastFour string, err error) {
	if len(pan) < 4 {
		return "", "", "", fmt.Errorf("PAN too short")
	}

	// Encrypt the real PAN
	encryptedPAN, err = t.encryptor.Encrypt(pan)
	if err != nil {
		return "", "", "", fmt.Errorf("PAN encryption failed: %w", err)
	}

	// Generate random token
	token, err = generateToken()
	if err != nil {
		return "", "", "", fmt.Errorf("token generation failed: %w", err)
	}

	lastFour = pan[len(pan)-4:]
	return token, encryptedPAN, lastFour, nil
}

// Detokenize - decrypt PAN from encrypted form
func (t *Tokenizer) Detokenize(encryptedPAN string) (string, error) {
	return t.encryptor.Decrypt(encryptedPAN)
}

func generateToken() (string, error) {
	b := make([]byte, 10)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("tok_%x", b), nil
}
