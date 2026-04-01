package http

import (
	"time"

	_ "github.com/BakhodiribnYashinibnMansur/XBank/docs/swagger"
	infraAuth "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/auth"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/config"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/handler"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/middleware"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	goredis "github.com/redis/go-redis/v9"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func NewRouter(
	userHandler *handler.UserHandler,
	authHandler *handler.AuthHandler,
	accountHandler *handler.AccountHandler,
	transferHandler *handler.TransferHandler,
	cardHandler *handler.CardHandler,
	beneficiaryHandler *handler.BeneficiaryHandler,
	exchangeHandler *handler.ExchangeHandler,
	kycHandler *handler.KYCHandler,
	fraudHandler *handler.FraudHandler,
	notificationHandler *handler.NotificationHandler,
	healthHandler *handler.HealthHandler,
	jwtService *infraAuth.JWTService,
	adminWhitelist *middleware.DynamicIPWhitelist,
	redisClient *goredis.Client,
	cfg *config.Config,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ErrorHandler: apperror.ErrorHandler,
	})

	// Middleware chain
	app.Use(middleware.RecoveryMiddleware())
	app.Use(middleware.RequestIDMiddleware())
	app.Use(middleware.HelmetMiddleware())
	app.Use(middleware.CSRFMiddleware())
	app.Use(middleware.TracingMiddleware())
	app.Use(middleware.CORSMiddleware(cfg.CORS.Origins()))
	app.Use(middleware.RateLimitMiddleware(cfg.RateLimit.MaxRequests, cfg.RateLimit.Window()))
	app.Use(middleware.MetricsMiddleware())
	app.Use(middleware.LoggerMiddleware())

	// Prometheus metrics endpoint
	prometheusHandler := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
	app.Get("/metrics", func(c *fiber.Ctx) error {
		prometheusHandler(c.Context())
		return nil
	})

	// Health probes
	app.Get("/health", healthHandler.Live)
	app.Get("/health/live", healthHandler.Live)
	app.Get("/health/ready", healthHandler.Ready)

	// API Documentation
	app.Get("/swagger/*", swagger.HandlerDefault)  // Swagger UI: http://localhost:3000/swagger/
	app.Get("/docs", handler.RedocHandler())        // ReDoc UI:   http://localhost:3000/docs

	v1 := app.Group("/api/v1")

	// Auth (public)
	auth := v1.Group("/auth")
	auth.Post("/register", userHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)
	auth.Post("/logout", authHandler.Logout)

	// Protected
	protected := v1.Group("", middleware.AuthMiddleware(jwtService), middleware.RLSMiddleware())
	protected.Post("/auth/logout-all", authHandler.LogoutAll)

	// Users: GET /api/v1/users/get?id=xxx
	users := protected.Group("/users")
	users.Get("/get", userHandler.GetByID)
	users.Post("/change-password", userHandler.ChangePassword)
	users.Get("/me/data-export", userHandler.ExportData)
	users.Delete("/me/delete", userHandler.DeleteAccount)

	// Accounts
	accounts := protected.Group("/accounts")
	accounts.Post("/create", accountHandler.Create)
	accounts.Get("/list", accountHandler.List)
	accounts.Get("/get", accountHandler.GetByID)
	accounts.Post("/deposit", accountHandler.Deposit)
	accounts.Post("/withdraw", accountHandler.Withdraw)
	accounts.Get("/history", accountHandler.History)

	// Account admin operations (ADMIN, TELLER only)
	accountAdmin := accounts.Group("", middleware.RequireRole("ADMIN", "TELLER"))
	accountAdmin.Post("/close", accountHandler.Close)

	// Transfers
	transfers := protected.Group("/transfers")
	transfers.Use(middleware.IdempotencyMiddleware(redisClient, 24*time.Hour))
	transfers.Post("/send", transferHandler.Send)
	transfers.Get("/get", transferHandler.GetByID)          // ?id=xxx
	transfers.Get("/list", transferHandler.ListByAccount)   // ?account_id=xxx
	transfers.Get("/history", transferHandler.History)      // ?id=xxx

	// Cards — RESTful design
	cards := protected.Group("/cards")
	cards.Post("/", cardHandler.Issue)              // POST /cards
	cards.Get("/", cardHandler.ByAccount)            // GET  /cards?account_id=xxx
	cards.Get("/:id", cardHandler.ByID)              // GET  /cards/:id
	cards.Post("/:id/activate", cardHandler.Activate)
	cards.Post("/:id/verify-pin", cardHandler.VerifyPIN)
	cards.Put("/:id/pin", cardHandler.ChangePIN)
	// Card admin operations (ADMIN only)
	cardAdmin := cards.Group("", middleware.RequireRole("ADMIN"))
	cardAdmin.Post("/:id/block", cardHandler.Block)
	cardAdmin.Post("/:id/unblock", cardHandler.Unblock)

	// Exchange rates (public: view, protected: convert, admin: set rates)
	currencies := v1.Group("/currencies")
	currencies.Get("/rate", exchangeHandler.GetRate)
	currencies.Get("/rates", exchangeHandler.ListRates)
	currencies.Post("/convert", middleware.AuthMiddleware(jwtService), exchangeHandler.Convert)
	currencies.Post("/rate", middleware.AuthMiddleware(jwtService), middleware.RequireRole("ADMIN"), exchangeHandler.UpsertRate)

	// Beneficiaries
	bens := protected.Group("/beneficiaries")
	bens.Post("/add", beneficiaryHandler.Add)
	bens.Get("/list", beneficiaryHandler.List)
	bens.Delete("/delete", beneficiaryHandler.Delete)

	// SSE Notifications
	notifications := protected.Group("/notifications")
	notifications.Get("/stream", notificationHandler.Stream)

	// KYC (customer)
	kycRoutes := protected.Group("/kyc")
	kycRoutes.Post("/submit", kycHandler.Submit)
	kycRoutes.Get("/status", kycHandler.Status)

	// Admin routes (ADMIN only + IP whitelist)
	admin := protected.Group("/admin", middleware.RequireRole("ADMIN"), adminWhitelist.Middleware())

	// KYC admin
	adminKYC := admin.Group("/kyc")
	adminKYC.Get("/pending", kycHandler.ListPending)
	adminKYC.Post("/approve", kycHandler.Approve)
	adminKYC.Post("/reject", kycHandler.Reject)

	// Fraud admin
	adminFraud := admin.Group("/fraud")
	adminFraud.Get("/flagged", fraudHandler.ListFlagged)
	adminFraud.Get("/check", fraudHandler.GetByTransfer)

	return app
}
