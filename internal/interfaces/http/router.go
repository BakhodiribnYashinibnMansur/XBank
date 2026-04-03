package http

import (
	"context"
	"time"

	_ "github.com/BakhodiribnYashinibnMansur/XBank/docs/swagger"
	infraAuth "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/auth"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/config"
	infraCrypto "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/crypto"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/handler"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/middleware"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	goredis "github.com/redis/go-redis/v9"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// challengeValidator - interface to avoid circular import with challenge service
type challengeValidator interface {
	ValidateToken(ctx context.Context, token, userID string) error
}

func NewRouter(
	userHandler *handler.UserHandler,
	authHandler *handler.AuthHandler,
	accountHandler *handler.AccountHandler,
	transferHandler *handler.TransferHandler,
	cardHandler *handler.CardHandler,
	cardExtHandler *handler.CardExtendedHandler,
	beneficiaryHandler *handler.BeneficiaryHandler,
	exchangeHandler *handler.ExchangeHandler,
	kycHandler *handler.KYCHandler,
	fraudHandler *handler.FraudHandler,
	notificationHandler *handler.NotificationHandler,
	contactHandler *handler.ContactHandler,
	healthHandler *handler.HealthHandler,
	auditHandler *handler.AuditHandler,
	reconHandler *handler.ReconciliationHandler,
	challengeHandler *handler.ChallengeHandler,
	totpHandler *handler.TOTPHandler,
	scheduledHandler *handler.ScheduledTransferHandler,
	jwtService *infraAuth.JWTService,
	adminWhitelist *middleware.DynamicIPWhitelist,
	hmacSigner *infraCrypto.HMACSigner,
	challengeService challengeValidator,
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

	// TOTP (public — called before JWT is issued)
	if totpHandler != nil {
		auth.Post("/totp/verify", totpHandler.VerifyLogin) // complete login with TOTP code
	}

	// Protected
	protected := v1.Group("", middleware.AuthMiddleware(jwtService), middleware.RLSMiddleware())
	protected.Post("/auth/logout-all", authHandler.LogoutAll)

	// TOTP management (protected — requires JWT)
	if totpHandler != nil {
		protected.Post("/auth/totp/setup", totpHandler.Setup)
		protected.Post("/auth/totp/confirm", totpHandler.ConfirmSetup)
		protected.Post("/auth/totp/disable", totpHandler.Disable)
	}

	// Challenge (step-up auth)
	if challengeHandler != nil {
		protected.Post("/auth/challenge/request", challengeHandler.Request)
		protected.Post("/auth/challenge/verify", challengeHandler.Verify)
	}

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

	// Transfers (HMAC signed + idempotent)
	transfers := protected.Group("/transfers")
	transfers.Use(middleware.HMACMiddleware(hmacSigner))
	transfers.Use(middleware.IdempotencyMiddleware(redisClient, 24*time.Hour))
	transfers.Post("/send", middleware.RequireChallenge(challengeService), transferHandler.Send)
	transfers.Get("/get", transferHandler.GetByID)          // ?id=xxx
	transfers.Get("/list", transferHandler.ListByAccount)   // ?account_id=xxx
	transfers.Get("/history", transferHandler.History)      // ?id=xxx

	// Scheduled transfers
	if scheduledHandler != nil {
		transfers.Post("/scheduled", scheduledHandler.Schedule)
		transfers.Post("/scheduled/cancel", scheduledHandler.Cancel)
		transfers.Get("/scheduled/get", scheduledHandler.GetByID)
		transfers.Get("/scheduled/list", scheduledHandler.List)
	}

	// Cards — RESTful design (HMAC signed)
	cards := protected.Group("/cards")
	cards.Use(middleware.HMACMiddleware(hmacSigner))
	cards.Post("/", cardHandler.Issue)              // POST /cards
	cards.Get("/", cardHandler.ByAccount)            // GET  /cards?account_id=xxx
	cards.Get("/:id", cardHandler.ByID)              // GET  /cards/:id
	cards.Post("/:id/activate", cardHandler.Activate)
	cards.Post("/:id/verify-pin", cardHandler.VerifyPIN)
	cards.Put("/:id/pin", cardHandler.ChangePIN)
	// Tokenization, 3DS, EMV, Holds
	if cardExtHandler != nil {
		cards.Post("/:id/tokenize", cardExtHandler.Tokenize)
		cards.Get("/:id/tokens", cardExtHandler.ListTokens)
		cards.Post("/tokens/revoke", cardExtHandler.RevokeToken)
		cards.Post("/:id/3ds/enroll", cardExtHandler.Enroll3DS)
		cards.Post("/:id/emv", cardExtHandler.SetEMV)
		cards.Get("/:id/holds", cardExtHandler.ListHolds)
		cards.Post("/holds", cardExtHandler.CreateHold)
		cards.Post("/holds/:id/capture", cardExtHandler.CaptureHold)
		cards.Post("/holds/:id/release", cardExtHandler.ReleaseHold)
	}

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

	// Contacts
	if contactHandler != nil {
		contacts := protected.Group("/contacts")
		contacts.Post("/add", contactHandler.Add)
		contacts.Get("/list", contactHandler.List)
		contacts.Delete("/delete", contactHandler.Delete)
	}

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

	// Audit admin
	if auditHandler != nil {
		admin.Get("/audit", auditHandler.List)
	}

	// Reconciliation admin
	if reconHandler != nil {
		adminRecon := admin.Group("/reconciliation")
		adminRecon.Get("/check", reconHandler.CheckAccount)
		adminRecon.Get("/check-all", reconHandler.CheckAllByUser)
	}

	return app
}
