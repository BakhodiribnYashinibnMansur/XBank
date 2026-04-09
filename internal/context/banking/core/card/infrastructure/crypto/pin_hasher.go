package crypto

import "golang.org/x/crypto/bcrypt"

// BcryptPINHasher implements domain.PINHasher using bcrypt.
type BcryptPINHasher struct {
	cost int
}

func NewBcryptPINHasher() *BcryptPINHasher {
	return &BcryptPINHasher{cost: bcrypt.DefaultCost}
}

func (h *BcryptPINHasher) Hash(pin string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pin), h.cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (h *BcryptPINHasher) Compare(hashedPIN, pin string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPIN), []byte(pin))
}
