package e2e_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/app"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/config"
	infraRedis "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/redis"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/eventbus"
	infraCrypto "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/security/crypto"
	infraAuth "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/security/jwt"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/outbox"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/BakhodiribnYashinibnMansur/XBank/test/integration/common/setup"
	"github.com/gofiber/fiber/v2"
)

var (
	testApp    *fiber.App
	hmacSecret string
)

func TestMain(m *testing.M) {
	if os.Getenv("E2E") == "" {
		fmt.Println("skipping E2E tests (set E2E=1 to run)")
		os.Exit(0)
	}

	logger.Init(true)
	os.Setenv("APP_ENV", "test")

	var code int
	func() {
		t := &testing.T{}

		// Start testcontainers
		pgc := setup.MustPostgres(t)
		defer pgc.Teardown()

		rdc := setup.MustRedis(t)
		defer func() { rdc.Client.Close(); rdc.Container.Terminate(context.Background()) }()

		// Generate temp JWT keys
		privateKeyPath, publicKeyPath := generateTempKeys(t)
		defer os.Remove(privateKeyPath)
		defer os.Remove(publicKeyPath)

		// HMAC secret (32 bytes hex)
		hmacBytes := make([]byte, 32)
		rand.Read(hmacBytes)
		hmacSecret = hex.EncodeToString(hmacBytes)

		// Config
		cfg := &config.Config{
			App: config.AppConfig{Name: "xbank-test", Port: 0},
			JWT: config.JWTConfig{
				PrivateKeyPath:   privateKeyPath,
				PublicKeyPath:    publicKeyPath,
				Issuer:           "xbank-test",
				Audience:         "xbank-test-client",
				AccessTTLMinutes: 30,
				RefreshTTLDays:   7,
			},
			RateLimit: config.RateLimitConfig{MaxRequests: 1000, WindowMinutes: 1},
			CORS:      config.CORSConfig{AllowedOrigins: []string{"http://localhost:3000"}},
			HMAC:      config.HMACConfig{MaxClockSkewMinutes: 5},
			Kafka: config.KafkaConfig{
				Topics: config.KafkaTopicsConfig{
					AccountOpened:     "test.account.opened",
					AccountCredited:   "test.account.credited",
					AccountDebited:    "test.account.debited",
					AccountFrozen:     "test.account.frozen",
					AccountClosed:     "test.account.closed",
					TransferCreated:   "test.transfer.created",
					TransferCompleted: "test.transfer.completed",
					TransferFailed:    "test.transfer.failed",
				},
			},
			Encryption: config.EncryptionConfig{
				HMACSecret: hmacSecret,
			},
		}

		// Infrastructure
		pool := pgc.Pool
		txManager := postgres.NewTxManager(pool)
		outboxRepo := outbox.NewRepository(pool)
		outboxPublisher := outbox.NewPublisher(outboxRepo)
		evBus := eventbus.New()

		jwtService, err := infraAuth.NewJWTService(
			cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath,
			cfg.JWT.Issuer, cfg.JWT.Audience,
			cfg.JWT.AccessTTL(), cfg.JWT.RefreshTTL(),
		)
		if err != nil {
			panic(fmt.Sprintf("JWT service: %v", err))
		}

		totpService := infraAuth.NewTOTPService(cfg.App.Name)
		sessionCache := infraRedis.NewSessionCache(rdc.Client)
		loginLimiter := infraRedis.NewLoginLimiter(rdc.Client, 5, 15*time.Minute, 15*time.Minute)
		challengeCache := infraRedis.NewChallengeCache(rdc.Client)

		// Card encryption (32-byte key)
		cardKey := make([]byte, 32)
		rand.Read(cardKey)
		cardKeyHex := hex.EncodeToString(cardKey)
		cardEncryptor, _ := infraCrypto.NewAESEncryptor(cardKeyHex)
		tokenizer := infraCrypto.NewTokenizer(cardEncryptor)

		hmacSigner, _ := infraCrypto.NewHMACSigner(hmacSecret, cfg.HMAC.MaxClockSkew())

		metrics.Register()

		// Bounded contexts
		bcs := app.NewDDDBoundedContexts(
			pool, txManager, outboxPublisher, evBus, cfg,
			jwtService, totpService,
			sessionCache, loginLimiter, challengeCache,
			cardEncryptor, tokenizer,
			nil, // auditLog (skip MongoDB)
		)

		// Build Fiber app (no MongoDB, no SSE)
		testApp = app.RegisterDDDRoutes(
			bcs, pool, nil, nil,
			jwtService, nil, hmacSigner,
			rdc.Client, cfg,
		)

		code = m.Run()
	}()
	os.Exit(code)
}

// generateTempKeys creates temporary ECDSA P-256 key files for testing.
func generateTempKeys(t *testing.T) (string, string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generating ECDSA key: %v", err)
	}

	// Private key
	privBytes, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshaling private key: %v", err)
	}
	privFile, _ := os.CreateTemp("", "jwt-private-*.pem")
	pem.Encode(privFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
	privFile.Close()

	// Public key
	pubBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshaling public key: %v", err)
	}
	pubFile, _ := os.CreateTemp("", "jwt-public-*.pem")
	pem.Encode(pubFile, &pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})
	pubFile.Close()

	return privFile.Name(), pubFile.Name()
}
