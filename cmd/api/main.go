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
	exchApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/exchange"
	cardApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/card"
	transferApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/transfer"
	userApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/user"
	infraAuth "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/auth"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/config"
	infraCrypto "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/crypto"
	infraKafka "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/kafka"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/metrics"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/tracing"
	infraMongo "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/mongodb"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/postgres"
	infraRedis "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/redis"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/handler"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.uber.org/zap"

	router "github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http"
)

func main() {
	logger.Init(true)
	defer logger.Sync()

	cfg := config.Load("config.yml")

	ctx := context.Background()

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
	_ = redisClient // TODO: wire into session cache and rate limiter

	jwtService, err := infraAuth.NewJWTService(
		cfg.JWT.PrivateKeyPath, cfg.JWT.PublicKeyPath,
		cfg.JWT.Issuer, cfg.JWT.Audience,
		cfg.JWT.AccessTTL(), cfg.JWT.RefreshTTL(),
	)
	if err != nil {
		logger.Log.Fatal("JWT service yaratib bo'lmadi", zap.Error(err))
	}

	txManager := postgres.NewTxManager(pool)

	userRepo := postgres.NewUserRepository(pool)
	sessionRepo := postgres.NewSessionRepository(pool)
	accountRepo := postgres.NewAccountRepository(pool)
	accountEventRepo := postgres.NewAccountEventRepository(pool)
	transferRepo := postgres.NewTransferRepository(pool)
	transferEventRepo := postgres.NewTransferEventRepository(pool)
	cardRepo := postgres.NewCardRepository(pool)

	// DB pool metrics collector (every 15s)
	poolCtx, poolCancel := context.WithCancel(ctx)
	metrics.StartDBPoolCollector(poolCtx, pool, 15*time.Second)

	// Kafka producer
	kafkaProducer := infraKafka.NewProducer(cfg.Kafka.Brokers)

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
	authService := authApp.NewService(userRepo, sessionRepo, jwtService)
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

	// Interfaces
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService)
	accountHandler := handler.NewAccountHandler(accountService)
	transferHandler := handler.NewTransferHandler(transferService)
	cardHandler := handler.NewCardHandler(cardService)

	benRepo := postgres.NewBeneficiaryRepository(pool)
	benService := benApp.NewService(benRepo)
	benHandler := handler.NewBeneficiaryHandler(benService)

	exchRepo := postgres.NewExchangeRepository(pool)
	exchService := exchApp.NewService(exchRepo)
	exchHandler := handler.NewExchangeHandler(exchService)

	kafkaBroker := ""
	if len(cfg.Kafka.Brokers) > 0 {
		kafkaBroker = cfg.Kafka.Brokers[0]
	}
	healthHandler := handler.NewHealthHandler(pool, mongoClient, kafkaBroker)

	app := router.NewRouter(userHandler, authHandler, accountHandler, transferHandler, cardHandler, benHandler, exchHandler, healthHandler, jwtService, cfg)

	// Graceful shutdown: wait for termination signal (Ctrl+C or docker stop)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

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

	// 1. Stop accepting new requests
	//    Wait for in-flight requests to complete
	if err := app.Shutdown(); err != nil {
		logger.Log.Error("Server shutdown xatolik", zap.Error(err))
	}

	// 2. Close Kafka producer
	if err := kafkaProducer.Close(); err != nil {
		logger.Log.Error("Kafka producer yopishda xatolik", zap.Error(err))
	}

	// 3. Close Redis
	if redisClient != nil {
		redisClient.Close()
	}

	// 4. Close MongoDB
	if mongoClient != nil {
		mongoClient.Disconnect(context.Background())
	}

	// 4. Flush tracing spans
	if err := shutdownTracer(context.Background()); err != nil {
		logger.Log.Error("Tracer shutdown xatolik", zap.Error(err))
	}

	// 5. Stop DB pool collector
	poolCancel()

	// 6. Close DB connections
	pool.Close()

	logger.Log.Info("Server toza yopildi")
}
