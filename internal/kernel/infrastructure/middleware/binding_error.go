package middleware

import (
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// BindingErrorMiddleware catches JSON parsing errors from BodyParser
// and returns a structured validation error response.
func BindingErrorMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if err == nil {
			return nil
		}

		// Check if it's a JSON syntax error
		errMsg := err.Error()
		if strings.Contains(errMsg, "invalid character") ||
			strings.Contains(errMsg, "unexpected end of JSON") ||
			strings.Contains(errMsg, "cannot unmarshal") {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"success":     false,
				"status_code": http.StatusBadRequest,
				"error": fiber.Map{
					"code":    "INVALID_JSON",
					"message": "Request body contains invalid JSON",
				},
			})
		}

		return err
	}
}
