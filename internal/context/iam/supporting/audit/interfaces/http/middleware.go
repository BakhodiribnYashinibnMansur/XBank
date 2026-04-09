package http

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/audit/application/command"
	"github.com/gofiber/fiber/v2"
)

// EndpointHistoryMiddleware records API endpoint access as fire-and-forget.
func EndpointHistoryMiddleware(svc *command.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		durationMs := int(time.Since(start).Milliseconds())

		userID, _ := c.Locals("user_id").(string)

		go svc.CreateEndpointHistory(context.Background(), command.CreateEndpointHistoryInput{
			Method:     c.Method(),
			Path:       c.Path(),
			StatusCode: c.Response().StatusCode(),
			UserID:     userID,
			IPAddress:  c.IP(),
			DurationMs: durationMs,
		})

		return err
	}
}
