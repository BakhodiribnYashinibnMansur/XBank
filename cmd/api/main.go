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

	accountApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/account"
	authApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/auth"
	cardApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/card"
	transferApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/transfer"
	userApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/user"
	infraAuth "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/auth"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/config"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/handler"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"go.uber.org/zap"

	router "github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http"
)

func main() {
	logger.Init(true)
	defer logger.Sync()

	cfg := config.Load("config.yml")

	// Infrastructure
	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, cfg.Database.URL)
	if err != nil {
		logger.Log.Fatal("PostgreSQL ga ulanib bo'lmadi", zap.Error(err))
	}

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
	transferRepo := postgres.NewTransferRepository(pool)
	cardRepo := postgres.NewCardRepository(pool)

	// Application
	userService := userApp.NewService(userRepo)
	authService := authApp.NewService(userRepo, sessionRepo, jwtService)
	accountService := accountApp.NewService(accountRepo, txManager)
	transferService := transferApp.NewService(transferRepo, accountRepo, txManager)
	cardService := cardApp.NewService(cardRepo)

	// Interfaces
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService)
	accountHandler := handler.NewAccountHandler(accountService)
	transferHandler := handler.NewTransferHandler(transferService)
	cardHandler := handler.NewCardHandler(cardService)

	app := router.NewRouter(userHandler, authHandler, accountHandler, transferHandler, cardHandler, jwtService, cfg)

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

	// 2. Close DB connections
	pool.Close()

	logger.Log.Info("Server toza yopildi")
}
