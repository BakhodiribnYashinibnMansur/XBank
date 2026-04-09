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
		BodyLimit:    10 * 1024 * 1024, // 10MB max request body (DoS prevention)
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
	registerAdminRoutes(admin, bcs)

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
	if bcs.Session != nil {
		auth.Post("/login", bcs.Session.Handler.Login)
		auth.Post("/refresh", bcs.Session.Handler.Refresh)
		auth.Post("/logout", bcs.Session.Handler.Logout)
		auth.Post("/totp/verify", bcs.Session.Handler.TOTPVerifyLogin)
	}

	// User BC — register
	if bcs.User != nil {
		auth.Post("/register", bcs.User.Handler.Register)
	}

	// Exchange BC — public rate queries
	if bcs.Exchange != nil {
		currencies := v1.Group("/currencies")
		currencies.Get("/rate", bcs.Exchange.Handler.GetRate)
		currencies.Get("/rates", bcs.Exchange.Handler.ListRates)
	}

	// Announcement BC — public active announcements
	if bcs.Announcement != nil {
		v1.Get("/announcements/active", bcs.Announcement.Handler.ListActive)
	}
}

// registerProtectedRoutes registers JWT-protected routes.
func registerProtectedRoutes(
	protected fiber.Router,
	bcs *DDDBoundedContexts,
	hmacSigner *infraCrypto.HMACSigner,
	redisClient *goredis.Client,
	cfg *config.Config,
) {
	if bcs.Session != nil {
		protected.Post("/auth/logout-all", bcs.Session.Handler.LogoutAll)
		protected.Post("/auth/totp/setup", bcs.Session.Handler.TOTPSetup)
		protected.Post("/auth/totp/confirm", bcs.Session.Handler.TOTPConfirmSetup)
		protected.Post("/auth/totp/disable", bcs.Session.Handler.TOTPDisable)
	}
	if bcs.Challenge != nil {
		bcs.Challenge.Handler.RegisterRoutes(protected)
	}
	if bcs.User != nil {
		bcs.User.Handler.RegisterRoutes(protected)
	}
	if bcs.Account != nil {
		bcs.Account.Handler.RegisterRoutes(protected)
	}

	// Transfer BC (HMAC + idempotency)
	if bcs.Transfer != nil {
		hmacGroup := protected.Group("")
		if hmacSigner != nil {
			hmacGroup.Use(middleware.HMACMiddleware(hmacSigner))
		}
		if redisClient != nil {
			hmacGroup.Use(middleware.IdempotencyMiddleware(redisClient, 24*time.Hour))
		}
		transferHTTP.RegisterRoutes(hmacGroup, bcs.Transfer.Handler, bcs.Transfer.ScheduledHandler)
	}

	// Card BC (HMAC)
	if bcs.Card != nil {
		cardHMACGroup := protected.Group("")
		if hmacSigner != nil {
			cardHMACGroup.Use(middleware.HMACMiddleware(hmacSigner))
		}
		bcs.Card.Handler.RegisterRoutes(cardHMACGroup)
		bcs.Card.ExtendedHandler.RegisterRoutes(cardHMACGroup)
	}
	if bcs.Beneficiary != nil {
		bcs.Beneficiary.Handler.RegisterRoutes(protected)
	}
	if bcs.Contact != nil {
		bcs.Contact.Handler.RegisterRoutes(protected)
	}
	if bcs.KYC != nil {
		bcs.KYC.Handler.RegisterRoutes(protected)
	}
	if bcs.Notification != nil {
		bcs.Notification.Handler.RegisterRoutes(protected)
	}
	if bcs.Exchange != nil {
		protected.Post("/currencies/convert", bcs.Exchange.Handler.Convert)
	}
	if bcs.FeatureFlag != nil {
		protected.Post("/flags/evaluate", bcs.FeatureFlag.Handler.Evaluate)
	}
}

// registerAdminRoutes registers admin-only routes.
func registerAdminRoutes(admin fiber.Router, bcs *DDDBoundedContexts) {
	if bcs.KYC != nil {
		bcs.KYC.Handler.RegisterAdminRoutes(admin)
	}
	if bcs.Fraud != nil {
		bcs.Fraud.Handler.RegisterRoutes(admin)
	}
	if bcs.Reconciliation != nil {
		bcs.Reconciliation.Handler.RegisterRoutes(admin)
	}
	if bcs.Authz != nil {
		bcs.Authz.Handler.RegisterRoutes(admin)
	}
	if bcs.Audit != nil {
		bcs.Audit.Handler.RegisterRoutes(admin)
	}
	if bcs.SystemError != nil {
		bcs.SystemError.Handler.RegisterRoutes(admin)
	}
	if bcs.ErrorCode != nil {
		bcs.ErrorCode.Handler.RegisterRoutes(admin)
	}
	if bcs.FeatureFlag != nil {
		bcs.FeatureFlag.Handler.RegisterAdminRoutes(admin)
	}
	if bcs.SiteSetting != nil {
		bcs.SiteSetting.Handler.RegisterRoutes(admin)
	}
	if bcs.Statistics != nil {
		bcs.Statistics.Handler.RegisterRoutes(admin)
	}
	if bcs.Translation != nil {
		bcs.Translation.Handler.RegisterRoutes(admin)
	}
	if bcs.Announcement != nil {
		bcs.Announcement.Handler.RegisterAdminRoutes(admin)
	}
	if bcs.Exchange != nil {
		admin.Post("/currencies/rate", bcs.Exchange.Handler.UpsertRate)
	}
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
