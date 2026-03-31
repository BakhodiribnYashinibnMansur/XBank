package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestIDMiddleware - assigns a unique ID to every request
//
// Why is this needed?
// 1. Debug: quickly find which request caused an error in logs
// 2. Tracing: filter all logs for a single request by ID
// 3. Client: if an error occurs, the client sends the ID to support for quick lookup
//
// How it works:
//   If the client sends an "X-Request-ID" header, that ID is used
//   Otherwise, a new UUID is generated
//   The ID is also set in the response header
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check for an ID from the client
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store in context (for use in handlers)
		c.Locals("request_id", requestID)

		// Set in the response header
		c.Set("X-Request-ID", requestID)

		return c.Next()
	}
}
