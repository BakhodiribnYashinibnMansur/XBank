package jwks

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// Handler returns a Fiber handler that serves the JWKS at /.well-known/jwks.json.
func (p *Provider) Handler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		c.Set("Cache-Control", "public, max-age=3600")
		return c.Status(http.StatusOK).JSON(p.KeySet())
	}
}
