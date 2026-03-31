package auth

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("noto'g'ri token")
	ErrExpiredToken = errors.New("token muddati tugagan")
)

// TokenClaims - data stored inside the JWT
type TokenClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// TokenPair - access + refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// JWTService - JWT token generation/validation using ES256 (ECDSA P-256)
//
// Difference from HS256:
//   HS256: single secret used for both signing and verifying (if the secret leaks, tokens can be forged)
//   ES256: private key for signing only | public key for verifying only (public key cannot forge tokens)
//
// Security advantages:
//   1. Private key stays on the auth server only; other services verify with the public key
//   2. Small key size but equivalent strength to RSA-3072 (256-bit vs 3072-bit)
//   3. Sign/verify is 10x faster than RSA
type JWTService struct {
	privateKey      *ecdsa.PrivateKey
	publicKey       *ecdsa.PublicKey
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	issuer          string // Who created the token (e.g. "xbank-api")
	audience        string // Who the token is for (e.g. "xbank-client")
}

// NewJWTService - reads keys from PEM files and creates the service
func NewJWTService(privateKeyPath, publicKeyPath, issuer, audience string, accessTTL, refreshTTL time.Duration) (*JWTService, error) {
	privateKey, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("private key o'qib bo'lmadi: %w", err)
	}

	publicKey, err := loadPublicKey(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("public key o'qib bo'lmadi: %w", err)
	}

	return &JWTService{
		privateKey:      privateKey,
		publicKey:       publicKey,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
		issuer:          issuer,
		audience:        audience,
	}, nil
}

// GenerateTokenPair - generates access (JWT ES256) + refresh (random) tokens
func (s *JWTService) GenerateTokenPair(userID, email string) (*TokenPair, error) {
	now := time.Now()

	claims := TokenClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Audience:  jwt.ClaimStrings{s.audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	// Sign with ES256 (private key is used)
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	accessToken, err := token.SignedString(s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("token sign xatolik: %w", err)
	}

	refreshToken, err := generateRandomToken()
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ValidateAccessToken - validates the token using the public key
func (s *JWTService) ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		// Verify the algorithm is ES256
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, ErrInvalidToken
		}
		return s.publicKey, nil // Verify with public key only
	},
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience),
		jwt.WithValidMethods([]string{"ES256"}),
	)

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// PublicKey - other services can obtain the public key for verification
func (s *JWTService) PublicKey() *ecdsa.PublicKey {
	return s.publicKey
}

func (s *JWTService) RefreshTokenTTL() time.Duration {
	return s.refreshTokenTTL
}

// HashToken - hashes the refresh token with SHA-256
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func generateRandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// loadPrivateKey - reads an ECDSA private key from a PEM file
func loadPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("PEM block topilmadi")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// loadPublicKey - reads an ECDSA public key from a PEM file
func loadPublicKey(path string) (*ecdsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("PEM block topilmadi")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	ecdsaKey, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("kalit ECDSA emas")
	}

	return ecdsaKey, nil
}
