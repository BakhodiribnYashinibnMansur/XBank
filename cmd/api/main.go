// @title           XBank API
// @version         1.0
// @description     XBank Banking API - DDD architecture with Go and Fiber
// @host            localhost:3000
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer {token}" to authorize
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	accountApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/account"
	authApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/auth"
	benApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/beneficiary"
	challengeApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/challenge"
	exchApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/exchange"
	fraudApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/fraud"
	kycApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/kyc"
	cardApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/card"
	contactApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/contact"
	reconApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/reconciliation"
	sagaApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/saga"
	transferApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/transfer"
	userApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/user"
	infraAuth "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/auth"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/config"
	infraCrypto "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/crypto"
	infraKafka "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/kafka"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/tracing"
	infraVault "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/vault"
	infraMongo "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/mongodb"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/postgres"
	infraRedis "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/redis"
	infraSSE "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/sse"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/handler"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/middleware"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.uber.org/zap"

	router "github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http"
)

func main() {
	logger.Init(true)
	defer logger.Sync()

	cfg := config.Load("config.yml")

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
				// Override config with Vault secrets (non-empty only)
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

	// Infrastructure
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
	if redisClient != nil {
		sessionCache = infraRedis.NewSessionCache(redisClient)
		loginLimiter = infraRedis.NewLoginLimiter(redisClient, 5, 15*time.Minute, 15*time.Minute)
		logger.Log.Info("Redis session cache + login limiter enabled")
	}

	jwtService, err := infraAuth.NewJWTService(
		cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath,
		cfg.JWT.Issuer, cfg.JWT.Audience,
		cfg.JWT.AccessTTL(), cfg.JWT.RefreshTTL(),
	)
	if err != nil {
		logger.Log.Fatal("JWT service yaratib bo'lmadi", zap.Error(err))
	}

	// TOTP service
	totpService := infraAuth.NewTOTPService(cfg.App.Name)

	txManager := postgres.NewTxManager(pool)

	userRepo := postgres.NewUserRepository(pool)
	sessionRepo := postgres.NewSessionRepository(pool)
	accountRepo := postgres.NewAccountRepository(pool)
	accountEventRepo := postgres.NewAccountEventRepository(pool)
	transferRepo := postgres.NewTransferRepository(pool)
	transferEventRepo := postgres.NewTransferEventRepository(pool)
	cardRepo := postgres.NewCardRepository(pool)
	contactRepo := postgres.NewContactRepository(pool)
	ledgerRepo := postgres.NewLedgerRepository(pool)

	// DB pool metrics collector (every 15s)
	poolCtx, poolCancel := context.WithCancel(ctx)
	metrics.StartDBPoolCollector(poolCtx, pool, 15*time.Second)

	// Kafka producer with DLQ fallback
	rawKafkaProducer := infraKafka.NewProducer(cfg.Kafka.Brokers)
	dlqRepo := postgres.NewDLQRepository(pool)
	dlqProducer := infraKafka.NewDLQProducer(rawKafkaProducer, dlqRepo)

	// Use DLQ producer as the event publisher (implements shared.EventPublisher)
	kafkaProducer := dlqProducer

	// Schema Registry — register Protobuf schemas at startup
	if cfg.Kafka.SchemaRegistryURL != "" {
		schemaRegistry := infraKafka.NewSchemaRegistry(cfg.Kafka.SchemaRegistryURL)
		go schemaRegistry.RegisterAllSchemas() // non-blocking, best-effort
		logger.Log.Info("Schema Registry enabled", zap.String("url", cfg.Kafka.SchemaRegistryURL))
	}

	// MongoDB audit log
	mongoClient, err := infraMongo.NewClient(ctx, cfg.MongoDB.URI)
	if err != nil {
		logger.Log.Warn("MongoDB unavailable, audit logging disabled", zap.Error(err))
	}
	var auditLog shared.AuditLog
	if mongoClient != nil {
		auditLog = infraMongo.NewAuditLog(mongoClient.Database(cfg.MongoDB.Database))
	}

	// Application
	userService := userApp.NewService(userRepo)
	authService := authApp.NewService(userRepo, sessionRepo, jwtService, totpService, sessionCache, loginLimiter)
	accountService := accountApp.NewService(accountRepo, accountEventRepo, txManager, kafkaProducer, cfg.Kafka.Topics, auditLog)
	transferService := transferApp.NewService(transferRepo, transferEventRepo, accountRepo, txManager, kafkaProducer, cfg.Kafka.Topics)

	var cardEncryptor *infraCrypto.AESEncryptor
	if cfg.Encryption.CardKey != "" {
		cardEncryptor, err = infraCrypto.NewAESEncryptor(cfg.Encryption.CardKey)
		if err != nil {
			logger.Log.Fatal("Card encryption key noto'g'ri", zap.Error(err))
		}
		logger.Log.Info("Card PAN encryption enabled")
	} else {
		logger.Log.Warn("Card PAN encryption disabled (CARD_ENCRYPTION_KEY not set)")
	}
	cardService := cardApp.NewService(cardRepo, cardEncryptor)

	// HMAC signer for request integrity (transfers, cards)
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

	// Interfaces
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService)
	accountHandler := handler.NewAccountHandler(accountService)
	transferHandler := handler.NewTransferHandler(transferService)
	cardHandler := handler.NewCardHandler(cardService)

	// Card extended: tokenization + holds + 3DS/EMV
	cardTokenRepo := postgres.NewCardTokenRepository(pool)
	cardHoldRepo := postgres.NewCardHoldRepository(pool)
	var tokenizer *infraCrypto.Tokenizer
	var tokenService *cardApp.TokenService
	if cardEncryptor != nil {
		tokenizer = infraCrypto.NewTokenizer(cardEncryptor)
		tokenService = cardApp.NewTokenService(cardTokenRepo, cardRepo, tokenizer)
		logger.Log.Info("Card tokenization enabled")
	}
	holdService := cardApp.NewHoldService(cardHoldRepo, cardRepo)
	cardExtHandler := handler.NewCardExtendedHandler(cardService, tokenService, holdService)

	benRepo := postgres.NewBeneficiaryRepository(pool)
	benService := benApp.NewService(benRepo)
	benHandler := handler.NewBeneficiaryHandler(benService)

	exchRepo := postgres.NewExchangeRepository(pool)
	exchService := exchApp.NewService(exchRepo)
	exchHandler := handler.NewExchangeHandler(exchService)

	kycRepo := postgres.NewKYCRepository(pool)
	kycService := kycApp.NewService(kycRepo)
	kycHandler := handler.NewKYCHandler(kycService)

	fraudRepo := postgres.NewFraudRepository(pool)
	fraudService := fraudApp.NewService(fraudRepo)
	fraudHandler := handler.NewFraudHandler(fraudService)

	// Transfer Saga (orchestrated multi-step transfer with compensations)
	transferSaga := sagaApp.NewTransferSaga(accountRepo, transferRepo, fraudRepo, ledgerRepo, txManager)
	_ = transferSaga // available for high-value transfer orchestration

	// Contacts
	contactService := contactApp.NewService(contactRepo)
	contactHandler := handler.NewContactHandler(contactService)

	// Audit read API (admin)
	var auditReader *infraMongo.AuditReader
	if mongoClient != nil {
		auditReader = infraMongo.NewAuditReader(mongoClient.Database(cfg.MongoDB.Database))
	}
	auditHandler := handler.NewAuditHandler(auditReader)

	// Reconciliation (admin)
	reconService := reconApp.NewService(accountRepo, ledgerRepo)
	reconHandler := handler.NewReconciliationHandler(reconService)

	// Challenge (step-up auth)
	challengeRepo := postgres.NewChallengeRepository(pool)
	var challengeCache *infraRedis.ChallengeCache
	if redisClient != nil {
		challengeCache = infraRedis.NewChallengeCache(redisClient)
	}
	challengeService := challengeApp.NewService(challengeRepo, userRepo, challengeCache)
	challengeHandler := handler.NewChallengeHandler(challengeService)
	totpHandler := handler.NewTOTPHandler(authService)

	// Scheduled transfers
	schedRepo := postgres.NewScheduledTransferRepository(pool)
	schedService := transferApp.NewScheduledService(schedRepo, transferService)
	scheduledHandler := handler.NewScheduledTransferHandler(schedService)

	sseHub := infraSSE.NewHub()
	notificationHandler := handler.NewNotificationHandler(sseHub)

	// Kafka Consumer — subscribe to domain events for SSE notifications
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

	kafkaBroker := ""
	if len(cfg.Kafka.Brokers) > 0 {
		kafkaBroker = cfg.Kafka.Brokers[0]
	}
	healthHandler := handler.NewHealthHandler(pool, mongoClient, kafkaBroker, redisClient)

	adminWhitelist := middleware.NewDynamicIPWhitelist(pool, 5*time.Minute)

	app := router.NewRouter(userHandler, authHandler, accountHandler, transferHandler, cardHandler, cardExtHandler, benHandler, exchHandler, kycHandler, fraudHandler, notificationHandler, contactHandler, healthHandler, auditHandler, reconHandler, challengeHandler, totpHandler, scheduledHandler, jwtService, adminWhitelist, hmacSigner, challengeService, redisClient, cfg)

	// Graceful shutdown: wait for termination signal (Ctrl+C or docker stop)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// DLQ retry background worker (every 30 seconds)
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

	// Scheduled transfers background worker (every 15 seconds)
	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-dlqCtx.Done():
				return
			case <-ticker.C:
				schedService.ExecuteDue(dlqCtx, 20)
			}
		}
	}()
	logger.Log.Info("Scheduled transfers worker started (every 15s)")

	// Start Kafka consumer (non-blocking, each topic gets its own goroutine)
	kafkaConsumer.Start(dlqCtx)
	logger.Log.Info("Kafka consumer started", zap.Int("topics", 8))

	// Start the server in a separate goroutine
	go func() {
		logger.Log.Info("Server ishga tushmoqda",
			zap.String("app", cfg.App.Name),
			zap.Int("port", cfg.App.Port),
		)
		if err := app.Listen(fmt.Sprintf(":%d", cfg.App.Port)); err != nil {
			logger.Log.Fatal("Server ishga tushmadi", zap.Error(err))
		}
	}()

	// Block here until a signal is received
	sig := <-quit
	logger.Log.Info("Shutdown signal qabul qilindi", zap.String("signal", sig.String()))

	// Shutdown with 30-second deadline for in-flight requests
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// 1. Stop accepting new requests, drain in-flight with timeout
	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		logger.Log.Error("Server shutdown xatolik (in-flight requests may be dropped)", zap.Error(err))
	}
	logger.Log.Info("HTTP server stopped")

	// 2. Stop background workers
	dlqCancel()

	// 3. Close Kafka consumer + producer
	if err := kafkaConsumer.Close(); err != nil {
		logger.Log.Error("Kafka consumer yopishda xatolik", zap.Error(err))
	}
	if err := kafkaProducer.Close(); err != nil {
		logger.Log.Error("Kafka producer yopishda xatolik", zap.Error(err))
	}

	// 4. Close Redis
	if redisClient != nil {
		redisClient.Close()
	}

	// 5. Close MongoDB
	if mongoClient != nil {
		mongoClient.Disconnect(shutdownCtx)
	}

	// 6. Flush tracing spans
	if err := shutdownTracer(shutdownCtx); err != nil {
		logger.Log.Error("Tracer shutdown xatolik", zap.Error(err))
	}

	// 7. Stop DB pool collector
	poolCancel()

	// 8. Close DB connections
	pool.Close()

	logger.Log.Info("Server toza yopildi")
}
