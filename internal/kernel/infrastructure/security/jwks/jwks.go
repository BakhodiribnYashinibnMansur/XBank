package jwks

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"sync"
)

// JWK represents a single JSON Web Key (RFC 7517).
type JWK struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
}

// JWKS is a JSON Web Key Set.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// Provider manages JWKS state and serves it via HTTP.
type Provider struct {
	mu   sync.RWMutex
	keys map[string]JWK // keyed by kid
}

// NewProvider creates an empty JWKS provider.
func NewProvider() *Provider {
	return &Provider{keys: make(map[string]JWK)}
}

// AddECDSAKey adds an ECDSA P-256 public key to the key set.
func (p *Provider) AddECDSAKey(kid string, pub *ecdsa.PublicKey) {
	if pub.Curve != elliptic.P256() {
		return // only P-256 supported for ES256
	}

	jwk := JWK{
		Kty: "EC",
		Crv: "P-256",
		X:   base64.RawURLEncoding.EncodeToString(pub.X.Bytes()),
		Y:   base64.RawURLEncoding.EncodeToString(pub.Y.Bytes()),
		Kid: kid,
		Use: "sig",
		Alg: "ES256",
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	p.keys[kid] = jwk
}

// RemoveKey removes a key by its kid.
func (p *Provider) RemoveKey(kid string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.keys, kid)
}

// JWKS returns the current JSON Web Key Set.
func (p *Provider) KeySet() JWKS {
	p.mu.RLock()
	defer p.mu.RUnlock()

	keys := make([]JWK, 0, len(p.keys))
	for _, k := range p.keys {
		keys = append(keys, k)
	}
	return JWKS{Keys: keys}
}
