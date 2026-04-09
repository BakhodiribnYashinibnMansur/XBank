package http

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Config holds HTTP server configuration.
type Config struct {
	Port            int
	AppName         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	ErrorHandler    fiber.ErrorHandler
}

// DefaultConfig returns sensible defaults.
func DefaultConfig(port int, appName string, errorHandler fiber.ErrorHandler) Config {
	return Config{
		Port:            port,
		AppName:         appName,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 30 * time.Second,
		ErrorHandler:    errorHandler,
	}
}

// NewFiber creates a configured Fiber app.
func NewFiber(cfg Config) *fiber.App {
	return fiber.New(fiber.Config{
		AppName:      cfg.AppName,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		ErrorHandler: cfg.ErrorHandler,
	})
}

// ListenAndServe starts the server and blocks until a shutdown signal is received.
// It performs graceful shutdown, waiting for in-flight requests to complete.
func ListenAndServe(app *fiber.App, cfg Config, onShutdown ...func(ctx context.Context)) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%d", cfg.Port)
		logger.Log.Info("HTTP server starting",
			zap.String("app", cfg.AppName),
			zap.Int("port", cfg.Port),
		)
		if err := app.Listen(addr); err != nil {
			logger.Log.Fatal("HTTP server failed", zap.Error(err))
		}
	}()

	sig := <-quit
	logger.Log.Info("Shutdown signal received", zap.String("signal", sig.String()))

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Log.Error("HTTP server shutdown error", zap.Error(err))
	}

	for _, fn := range onShutdown {
		fn(ctx)
	}

	logger.Log.Info("HTTP server stopped")
}
