package app

import (
	"bufio"
	"fmt"
	"time"

	_ "github.com/BakhodiribnYashinibnMansur/XBank/docs/swagger"
	transferHTTP "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/interfaces/http"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/config"
	infraCrypto "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/security/crypto"
	infraAuth "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/security/jwt"
	infraMongo "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/mongodb"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/middleware"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/sse"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	goredis "github.com/redis/go-redis/v9"
	"github.com/valyala/fasthttp/fasthttpadaptor"
	"go.mongodb.org/mongo-driver/mongo"
)

// challengeValidator avoids circular import with challenge service.
type challengeValidator interface {
	ValidateToken(ctx interface{ Context() interface{} }, token, userID string) error
}

// RegisterDDDRoutes builds the Fiber app and registers all BC routes.
func RegisterDDDRoutes(
	bcs *DDDBoundedContexts,
	pool *pgxpool.Pool,
	mongoClient *mongo.Client,
	auditReader *infraMongo.AuditReader,
	sseHub *sse.Hub,
	jwtService *infraAuth.JWTService,
	adminWhitelist *middleware.DynamicIPWhitelist,
	hmacSigner *infraCrypto.HMACSigner,
	redisClient *goredis.Client,
	cfg *config.Config,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ErrorHandler: apperror.ErrorHandler,
	})

	// Global middleware chain
	app.Use(middleware.RecoveryMiddleware())
	app.Use(middleware.RequestIDMiddleware())
	app.Use(middleware.HelmetMiddleware())
	app.Use(middleware.CSRFMiddleware())
	app.Use(middleware.TracingMiddleware())
	app.Use(middleware.CORSMiddleware(cfg.CORS.Origins()))
	app.Use(middleware.RateLimitMiddleware(cfg.RateLimit.MaxRequests, cfg.RateLimit.Window()))
	app.Use(middleware.MetricsMiddleware())
	app.Use(middleware.LoggerMiddleware())

	// Infrastructure endpoints
	prometheusHandler := fasthttpadaptor.NewFastHTTPHandler(promhttp.Handler())
	app.Get("/metrics", func(c *fiber.Ctx) error {
		prometheusHandler(c.Context())
		return nil
	})

	// Health probes
	kafkaBroker := ""
	if len(cfg.Kafka.Brokers) > 0 {
		kafkaBroker = cfg.Kafka.Brokers[0]
	}
	healthHandler := NewHealthHandler(pool, mongoClient, kafkaBroker, redisClient)
	app.Get("/health", healthHandler.Live)
	app.Get("/health/live", healthHandler.Live)
	app.Get("/health/ready", healthHandler.Ready)

	// API Documentation
	app.Get("/swagger/*", swagger.HandlerDefault)
	app.Get("/docs", RedocHandler())

	v1 := app.Group("/api/v1")

	// ── Public routes ──────────────────────────────────────────
	registerPublicRoutes(v1, bcs)

	// ── Protected routes (JWT required) ────────────────────────
	protected := v1.Group("", middleware.AuthMiddleware(jwtService), middleware.RLSMiddleware())
	registerProtectedRoutes(protected, bcs, hmacSigner, redisClient, cfg)

	// ── Admin routes (ADMIN role + IP whitelist) ───────────────
	admin := protected.Group("/admin", middleware.RequireRole("ADMIN"), adminWhitelist.Middleware())
	registerAdminRoutes(admin, bcs, auditReader)

	// ── SSE notification stream ────────────────────────────────
	if sseHub != nil {
		protected.Get("/notifications/stream", sseStreamHandler(sseHub))
	}

	return app
}

// registerPublicRoutes registers unauthenticated routes.
func registerPublicRoutes(v1 fiber.Router, bcs *DDDBoundedContexts) {
	auth := v1.Group("/auth")

	// Session BC — login, refresh, logout (public part)
	auth.Post("/login", bcs.Session.Handler.Login)
	auth.Post("/refresh", bcs.Session.Handler.Refresh)
	auth.Post("/logout", bcs.Session.Handler.Logout)
	auth.Post("/totp/verify", bcs.Session.Handler.TOTPVerifyLogin)

	// User BC — register
	auth.Post("/register", bcs.User.Handler.Register)

	// Exchange BC — public rate queries
	currencies := v1.Group("/currencies")
	currencies.Get("/rate", bcs.Exchange.Handler.GetRate)
	currencies.Get("/rates", bcs.Exchange.Handler.ListRates)

	// Announcement BC — public active announcements
	v1.Get("/announcements/active", bcs.Announcement.Handler.ListActive)
}

// registerProtectedRoutes registers JWT-protected routes.
func registerProtectedRoutes(
	protected fiber.Router,
	bcs *DDDBoundedContexts,
	hmacSigner *infraCrypto.HMACSigner,
	redisClient *goredis.Client,
	cfg *config.Config,
) {
	// Session BC — logout-all, TOTP management
	protected.Post("/auth/logout-all", bcs.Session.Handler.LogoutAll)
	protected.Post("/auth/totp/setup", bcs.Session.Handler.TOTPSetup)
	protected.Post("/auth/totp/confirm", bcs.Session.Handler.TOTPConfirmSetup)
	protected.Post("/auth/totp/disable", bcs.Session.Handler.TOTPDisable)

	// Challenge BC — step-up auth
	bcs.Challenge.Handler.RegisterRoutes(protected)

	// User BC
	bcs.User.Handler.RegisterRoutes(protected)

	// Account BC
	bcs.Account.Handler.RegisterRoutes(protected)

	// Transfer BC (HMAC + idempotency)
	hmacGroup := protected.Group("")
	if hmacSigner != nil {
		hmacGroup.Use(middleware.HMACMiddleware(hmacSigner))
	}
	if redisClient != nil {
		hmacGroup.Use(middleware.IdempotencyMiddleware(redisClient, 24*time.Hour))
	}
	transferHTTP.RegisterRoutes(hmacGroup, bcs.Transfer.Handler, bcs.Transfer.ScheduledHandler)

	// Card BC (HMAC)
	cardHMACGroup := protected.Group("")
	if hmacSigner != nil {
		cardHMACGroup.Use(middleware.HMACMiddleware(hmacSigner))
	}
	bcs.Card.Handler.RegisterRoutes(cardHMACGroup)
	bcs.Card.ExtendedHandler.RegisterRoutes(cardHMACGroup)

	// Beneficiary BC
	bcs.Beneficiary.Handler.RegisterRoutes(protected)

	// Contact BC
	bcs.Contact.Handler.RegisterRoutes(protected)

	// KYC BC (customer routes)
	bcs.KYC.Handler.RegisterRoutes(protected)

	// Notification BC (CQRS)
	bcs.Notification.Handler.RegisterRoutes(protected)

	// Exchange BC — protected convert
	protected.Post("/currencies/convert", bcs.Exchange.Handler.Convert)

	// Feature Flag BC — evaluate (public to authenticated users)
	protected.Post("/flags/evaluate", bcs.FeatureFlag.Handler.Evaluate)
}

// registerAdminRoutes registers admin-only routes.
func registerAdminRoutes(admin fiber.Router, bcs *DDDBoundedContexts, auditReader *infraMongo.AuditReader) {
	// KYC BC — admin
	bcs.KYC.Handler.RegisterAdminRoutes(admin)

	// Fraud BC
	bcs.Fraud.Handler.RegisterRoutes(admin)

	// Reconciliation BC
	bcs.Reconciliation.Handler.RegisterRoutes(admin)

	// Audit handler (infrastructure)
	auditHandler := NewAuditHandler(auditReader)
	admin.Get("/audit", auditHandler.List)

	// System Error BC
	bcs.SystemError.Handler.RegisterRoutes(admin)

	// Error Code BC
	bcs.ErrorCode.Handler.RegisterRoutes(admin)

	// Feature Flag BC — admin CRUD
	bcs.FeatureFlag.Handler.RegisterAdminRoutes(admin)

	// Site Setting BC
	bcs.SiteSetting.Handler.RegisterRoutes(admin)

	// Statistics BC
	bcs.Statistics.Handler.RegisterRoutes(admin)

	// Translation BC
	bcs.Translation.Handler.RegisterRoutes(admin)

	// Announcement BC — admin CRUD
	bcs.Announcement.Handler.RegisterAdminRoutes(admin)

	// Exchange BC — admin rate management
	admin.Post("/currencies/rate", bcs.Exchange.Handler.UpsertRate)
}

// sseStreamHandler returns a Fiber handler for SSE notification streaming.
func sseStreamHandler(hub *sse.Hub) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return apperror.ErrUnauthorized
		}

		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")

		c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
			ch := hub.Subscribe(userID)
			defer hub.Unsubscribe(userID, ch)

			fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"ok\"}\n\n")
			w.Flush()

			for data := range ch {
				fmt.Fprintf(w, "event: notification\ndata: %s\n\n", data)
				if err := w.Flush(); err != nil {
					return
				}
			}
		})

		return nil
	}
}
