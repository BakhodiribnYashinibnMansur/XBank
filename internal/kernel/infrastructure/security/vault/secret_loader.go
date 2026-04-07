package vault

import (
	"context"
	"fmt"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.uber.org/zap"
)

// SecretLoader reads application secrets from Vault and returns them
// as a flat map that can be applied to the Config struct.
// Secret paths in Vault:
//
//	secret/xbank/database   → database_url
//	secret/xbank/encryption → card_key, hmac_secret
//	secret/xbank/jwt        → private_key, public_key
//	secret/xbank/redis      → url
//	secret/xbank/mongodb    → uri
type SecretLoader struct {
	client *Client
	prefix string // path prefix, e.g. "xbank"
}

// NewSecretLoader creates a loader with the given Vault client and path prefix.
func NewSecretLoader(client *Client, prefix string) *SecretLoader {
	if prefix == "" {
		prefix = "xbank"
	}
	return &SecretLoader{client: client, prefix: prefix}
}

// Secrets holds all secrets loaded from Vault.
type Secrets struct {
	DatabaseURL   string
	RedisURL      string
	MongoDBURI    string
	CardKey       string
	HMACSecret    string
	JWTPrivateKey string
	JWTPublicKey  string
}

// Load reads all application secrets from Vault.
// Missing secrets are returned as empty strings (non-fatal).
func (l *SecretLoader) Load(ctx context.Context) (*Secrets, error) {
	s := &Secrets{}

	// Database credentials
	if data, err := l.client.GetSecret(ctx, l.path("database")); err == nil {
		s.DatabaseURL = data["url"]
		logger.Log.Info("Vault: loaded database secret")
	} else {
		logger.Log.Warn("Vault: database secret not found", zap.Error(err))
	}

	// Encryption keys
	if data, err := l.client.GetSecret(ctx, l.path("encryption")); err == nil {
		s.CardKey = data["card_key"]
		s.HMACSecret = data["hmac_secret"]
		logger.Log.Info("Vault: loaded encryption secrets")
	} else {
		logger.Log.Warn("Vault: encryption secrets not found", zap.Error(err))
	}

	// JWT signing keys
	if data, err := l.client.GetSecret(ctx, l.path("jwt")); err == nil {
		s.JWTPrivateKey = data["private_key"]
		s.JWTPublicKey = data["public_key"]
		logger.Log.Info("Vault: loaded JWT keys")
	} else {
		logger.Log.Warn("Vault: JWT keys not found", zap.Error(err))
	}

	// Redis
	if data, err := l.client.GetSecret(ctx, l.path("redis")); err == nil {
		s.RedisURL = data["url"]
		logger.Log.Info("Vault: loaded Redis secret")
	} else {
		logger.Log.Warn("Vault: Redis secret not found", zap.Error(err))
	}

	// MongoDB
	if data, err := l.client.GetSecret(ctx, l.path("mongodb")); err == nil {
		s.MongoDBURI = data["uri"]
		logger.Log.Info("Vault: loaded MongoDB secret")
	} else {
		logger.Log.Warn("Vault: MongoDB secret not found", zap.Error(err))
	}

	return s, nil
}

func (l *SecretLoader) path(name string) string {
	return fmt.Sprintf("%s/%s", l.prefix, name)
}
