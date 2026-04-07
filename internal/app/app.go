package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/config"
	infraCrypto "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/security/crypto"
	infraKafka "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/kafka"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/tracing"
	infraVault "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/security/vault"
	infraMongo "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/mongodb"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/postgres"
	infraRedis "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/redis"
	infraSSE "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/sse"
	infraAuth "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/security/jwt"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/eventbus"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/middleware"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.uber.org/zap"
)

// Run bootstraps all infrastructure, wires dependencies, and starts the server.
func Run(cfg *config.Config) {
	ctx := context.Background()

	// Vault — load secrets if enabled (overrides ENV-based config)
	if cfg.Vault.Enabled && cfg.Vault.Token != "" {
		vaultClient, vaultErr := infraVault.NewClient(ctx, infraVault.Config{
			Address:   cfg.Vault.Address,
			Token:     cfg.Vault.Token,
			MountPath: cfg.Vault.MountPath,
		})
		if vaultErr != nil {
			logger.Log.Warn("Vault unavailable, falling back to ENV secrets", zap.Error(vaultErr))
		} else {
			defer vaultClient.Close()
			loader := infraVault.NewSecretLoader(vaultClient, "xbank")
			secrets, loadErr := loader.Load(ctx)
			if loadErr != nil {
				logger.Log.Warn("Vault secret load failed", zap.Error(loadErr))
			} else {
				if secrets.DatabaseURL != "" {
					cfg.Database.URL = secrets.DatabaseURL
				}
				if secrets.RedisURL != "" {
					cfg.Redis.URL = secrets.RedisURL
				}
				if secrets.MongoDBURI != "" {
					cfg.MongoDB.URI = secrets.MongoDBURI
				}
				if secrets.CardKey != "" {
					cfg.Encryption.CardKey = secrets.CardKey
				}
				if secrets.HMACSecret != "" {
					cfg.Encryption.HMACSecret = secrets.HMACSecret
				}
				logger.Log.Info("Vault secrets applied to config")
			}
		}
	}

	// Tracing (OpenTelemetry → Jaeger)
	shutdownTracer, err := tracing.Init(ctx, tracing.Config{
		Endpoint:    cfg.Jaeger.Endpoint,
		ServiceName: cfg.App.Name,
		Enabled:     cfg.Jaeger.Enabled,
	})
	if err != nil {
		logger.Log.Fatal("Tracing ishga tushmadi", zap.Error(err))
	}

	// Metrics
	metrics.Register()

	// ── Infrastructure ─────────────────────────────────────────
	pool, err := postgres.NewPool(ctx, cfg.Database.URL)
	if err != nil {
		logger.Log.Fatal("PostgreSQL ga ulanib bo'lmadi", zap.Error(err))
	}

	redisClient, err := infraRedis.NewClient(ctx, cfg.Redis.URL)
	if err != nil {
		logger.Log.Warn("Redis unavailable", zap.Error(err))
	}
	var sessionCache *infraRedis.SessionCache
	var loginLimiter *infraRedis.LoginLimiter
	var challengeCache *infraRedis.ChallengeCache
	if redisClient != nil {
		sessionCache = infraRedis.NewSessionCache(redisClient)
		loginLimiter = infraRedis.NewLoginLimiter(redisClient, 5, 15*time.Minute, 15*time.Minute)
		challengeCache = infraRedis.NewChallengeCache(redisClient)
		logger.Log.Info("Redis session cache + login limiter + challenge cache enabled")
	}

	jwtService, err := infraAuth.NewJWTService(
		cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath,
		cfg.JWT.Issuer, cfg.JWT.Audience,
		cfg.JWT.AccessTTL(), cfg.JWT.RefreshTTL(),
	)
	if err != nil {
		logger.Log.Fatal("JWT service yaratib bo'lmadi", zap.Error(err))
	}

	totpService := infraAuth.NewTOTPService(cfg.App.Name)
	txManager := postgres.NewTxManager(pool)

	// DB pool metrics collector (every 15s)
	poolCtx, poolCancel := context.WithCancel(ctx)
	metrics.StartDBPoolCollector(poolCtx, pool, 15*time.Second)

	// Kafka producer with DLQ fallback
	rawKafkaProducer := infraKafka.NewProducer(cfg.Kafka.Brokers)
	dlqRepo := postgres.NewDLQRepository(pool)
	dlqProducer := infraKafka.NewDLQProducer(rawKafkaProducer, dlqRepo)
	kafkaProducer := dlqProducer

	// Schema Registry
	if cfg.Kafka.SchemaRegistryURL != "" {
		schemaRegistry := infraKafka.NewSchemaRegistry(cfg.Kafka.SchemaRegistryURL)
		go schemaRegistry.RegisterAllSchemas()
		logger.Log.Info("Schema Registry enabled", zap.String("url", cfg.Kafka.SchemaRegistryURL))
	}

	// MongoDB audit log
	mongoClient, err := infraMongo.NewClient(ctx, cfg.MongoDB.URI)
	if err != nil {
		logger.Log.Warn("MongoDB unavailable, audit logging disabled", zap.Error(err))
	}
	var auditLog domain.AuditLog
	var auditReader *infraMongo.AuditReader
	if mongoClient != nil {
		auditLog = infraMongo.NewAuditLog(mongoClient.Database(cfg.MongoDB.Database))
		auditReader = infraMongo.NewAuditReader(mongoClient.Database(cfg.MongoDB.Database))
	}

	// Card encryption
	var cardEncryptor *infraCrypto.AESEncryptor
	var tokenizer *infraCrypto.Tokenizer
	if cfg.Encryption.CardKey != "" {
		cardEncryptor, err = infraCrypto.NewAESEncryptor(cfg.Encryption.CardKey)
		if err != nil {
			logger.Log.Fatal("Card encryption key noto'g'ri", zap.Error(err))
		}
		tokenizer = infraCrypto.NewTokenizer(cardEncryptor)
		logger.Log.Info("Card PAN encryption + tokenization enabled")
	} else {
		logger.Log.Warn("Card PAN encryption disabled (CARD_ENCRYPTION_KEY not set)")
	}

	// HMAC signer
	var hmacSigner *infraCrypto.HMACSigner
	if cfg.Encryption.HMACSecret != "" {
		hmacSigner, err = infraCrypto.NewHMACSigner(cfg.Encryption.HMACSecret, cfg.HMAC.MaxClockSkew())
		if err != nil {
			logger.Log.Fatal("HMAC secret noto'g'ri", zap.Error(err))
		}
		logger.Log.Info("HMAC request signing enabled")
	} else {
		logger.Log.Warn("HMAC request signing disabled (HMAC_SECRET not set)")
	}

	// Event bus (in-memory)
	evBus := eventbus.New()

	// SSE hub
	sseHub := infraSSE.NewHub()

	// ── DDD Bounded Contexts ───────────────────────────────────
	bcs := NewDDDBoundedContexts(
		pool, txManager, kafkaProducer, evBus, cfg,
		jwtService, totpService,
		sessionCache, loginLimiter, challengeCache,
		cardEncryptor, tokenizer,
		auditLog,
	)

	// Admin IP whitelist
	adminWhitelist := middleware.NewDynamicIPWhitelist(pool, 5*time.Minute)

	// ── HTTP Router (DDD routes) ───────────────────────────────
	fiberApp := RegisterDDDRoutes(
		bcs, pool, mongoClient, auditReader, sseHub,
		jwtService, adminWhitelist, hmacSigner,
		redisClient, cfg,
	)

	// ── Kafka Consumer ─────────────────────────────────────────
	kafkaConsumer := infraKafka.NewConsumer(cfg.Kafka.Brokers, "xbank-api-consumer")
	accountEventHandler := infraKafka.NewAccountEventHandler(sseHub)
	transferEventHandler := infraKafka.NewTransferEventHandler(sseHub)

	kafkaConsumer.Subscribe(cfg.Kafka.Topics.AccountOpened, accountEventHandler)
	kafkaConsumer.Subscribe(cfg.Kafka.Topics.AccountCredited, accountEventHandler)
	kafkaConsumer.Subscribe(cfg.Kafka.Topics.AccountDebited, accountEventHandler)
	kafkaConsumer.Subscribe(cfg.Kafka.Topics.AccountFrozen, accountEventHandler)
	kafkaConsumer.Subscribe(cfg.Kafka.Topics.AccountClosed, accountEventHandler)
	kafkaConsumer.Subscribe(cfg.Kafka.Topics.TransferCreated, transferEventHandler)
	kafkaConsumer.Subscribe(cfg.Kafka.Topics.TransferCompleted, transferEventHandler)
	kafkaConsumer.Subscribe(cfg.Kafka.Topics.TransferFailed, transferEventHandler)

	// ── Graceful shutdown ──────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Background: DLQ retry worker (every 30s)
	dlqCtx, dlqCancel := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-dlqCtx.Done():
				return
			case <-ticker.C:
				dlqProducer.RetryPending(dlqCtx, 50)
			}
		}
	}()
	logger.Log.Info("DLQ retry worker started (every 30s)")

	// Background: scheduled transfers worker (every 15s)
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-dlqCtx.Done():
				return
			case <-ticker.C:
				bcs.Transfer.ScheduledService.ExecuteDue(dlqCtx, 20)
			}
		}
	}()
	logger.Log.Info("Scheduled transfers worker started (every 15s)")

	kafkaConsumer.Start(dlqCtx)
	logger.Log.Info("Kafka consumer started", zap.Int("topics", 8))

	// ── Start server ───────────────────────────────────────────
	go func() {
		logger.Log.Info("Server ishga tushmoqda",
			zap.String("app", cfg.App.Name),
			zap.Int("port", cfg.App.Port),
		)
		if err := fiberApp.Listen(fmt.Sprintf(":%d", cfg.App.Port)); err != nil {
			logger.Log.Fatal("Server ishga tushmadi", zap.Error(err))
		}
	}()

	sig := <-quit
	logger.Log.Info("Shutdown signal qabul qilindi", zap.String("signal", sig.String()))

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := fiberApp.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Log.Error("Server shutdown xatolik", zap.Error(err))
	}
	logger.Log.Info("HTTP server stopped")

	dlqCancel()

	if err := kafkaConsumer.Close(); err != nil {
		logger.Log.Error("Kafka consumer yopishda xatolik", zap.Error(err))
	}
	if err := kafkaProducer.Close(); err != nil {
		logger.Log.Error("Kafka producer yopishda xatolik", zap.Error(err))
	}

	if redisClient != nil {
		redisClient.Close()
	}
	if mongoClient != nil {
		mongoClient.Disconnect(shutdownCtx)
	}
	if err := shutdownTracer(shutdownCtx); err != nil {
		logger.Log.Error("Tracer shutdown xatolik", zap.Error(err))
	}

	poolCancel()
	pool.Close()

	logger.Log.Info("Server toza yopildi")
}
