package middleware

import (
	"net/http"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func BenchmarkHelmetMiddleware(b *testing.B) {
	app := fiber.New()
	app.Use(HelmetMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		app.Test(req, -1)
	}
}

func BenchmarkRequestIDMiddleware(b *testing.B) {
	app := fiber.New()
	app.Use(RequestIDMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		app.Test(req, -1)
	}
}

func BenchmarkRecoveryMiddleware(b *testing.B) {
	app := fiber.New()
	app.Use(RecoveryMiddleware())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		app.Test(req, -1)
	}
}

func BenchmarkRateLimitMiddleware(b *testing.B) {
	app := fiber.New()
	app.Use(RateLimitMiddleware(1000000, 1*time.Minute)) // high limit to avoid blocking
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		app.Test(req, -1)
	}
}

func BenchmarkFullMiddlewareStack(b *testing.B) {
	app := fiber.New()
	app.Use(RecoveryMiddleware())
	app.Use(RequestIDMiddleware())
	app.Use(HelmetMiddleware())
	app.Use(RateLimitMiddleware(1000000, 1*time.Minute))
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		app.Test(req, -1)
	}
}

func BenchmarkAuthMiddleware_ValidToken(b *testing.B) {
	jwtService := newTestJWTService(&testing.T{})
	// This is a workaround — in real benchmarks, set up once
	app := fiber.New()
	app.Get("/test", AuthMiddleware(jwtService), func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	pair, _ := jwtService.GenerateTokenPair("user-1", "user@test.com", "CUSTOMER", "0.0.0.0")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
		app.Test(req, -1)
	}
}
