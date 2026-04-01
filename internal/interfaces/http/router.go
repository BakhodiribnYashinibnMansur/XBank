package http

import (
	_ "github.com/BakhodiribnYashinibnMansur/XBank/docs/swagger"
	infraAuth "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/auth"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/config"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/handler"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/middleware"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

func NewRouter(
	userHandler *handler.UserHandler,
	authHandler *handler.AuthHandler,
	accountHandler *handler.AccountHandler,
	transferHandler *handler.TransferHandler,
	cardHandler *handler.CardHandler,
	jwtService *infraAuth.JWTService,
	cfg *config.Config,
) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      cfg.App.Name,
		ErrorHandler: apperror.ErrorHandler,
	})

	// Middleware chain
	app.Use(middleware.RecoveryMiddleware())
	app.Use(middleware.RequestIDMiddleware())
	app.Use(middleware.CORSMiddleware(cfg.CORS.Origins()))
	app.Use(middleware.RateLimitMiddleware(cfg.RateLimit.MaxRequests, cfg.RateLimit.Window()))
	app.Use(middleware.LoggerMiddleware())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Swagger UI: http://localhost:3000/swagger/
	app.Get("/swagger/*", swagger.HandlerDefault)

	v1 := app.Group("/api/v1")

	// Auth (public)
	auth := v1.Group("/auth")
	auth.Post("/register", userHandler.Register)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)
	auth.Post("/logout", authHandler.Logout)

	// Protected
	protected := v1.Group("", middleware.AuthMiddleware(jwtService))
	protected.Post("/auth/logout-all", authHandler.LogoutAll)

	// Users: GET /api/v1/users/get?id=xxx
	users := protected.Group("/users")
	users.Get("/get", userHandler.GetByID)

	// Accounts: static paths, ID comes from query/body
	accounts := protected.Group("/accounts")
	accounts.Post("/create", accountHandler.Create)
	accounts.Get("/list", accountHandler.List)
	accounts.Get("/get", accountHandler.GetByID)       // ?id=xxx
	accounts.Post("/deposit", accountHandler.Deposit)   // body: {account_id, amount}
	accounts.Post("/withdraw", accountHandler.Withdraw) // body: {account_id, amount}
	accounts.Post("/close", accountHandler.Close)       // body: {account_id}
	accounts.Get("/history", accountHandler.History)    // ?id=xxx

	// Transfers
	transfers := protected.Group("/transfers")
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
	cards.Post("/:id/block", cardHandler.Block)
	cards.Post("/:id/unblock", cardHandler.Unblock)

	return app
}
