package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSMiddleware - Cross-Origin Resource Sharing
//
// What is CORS?
// A browser security policy: a frontend (e.g. localhost:5173)
// cannot send requests directly to a backend (localhost:3000).
// CORS middleware tells the backend to "allow requests from these origins".
func CORSMiddleware(allowedOrigins string) fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     "GET,POST,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Request-ID",
		AllowCredentials: true,
	})
}
