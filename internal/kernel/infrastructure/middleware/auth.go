package middleware

import (
	"strings"

	infraAuth "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/security/jwt"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware - validates the JWT token
// Protected routes can only be accessed with a valid token
//
// Request header: Authorization: Bearer <access_token>
//
// If the token is valid, stores user_id and email in the context
// Subsequent handlers can retrieve them via c.Locals("user_id")
func AuthMiddleware(jwtService *infraAuth.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return apperror.ErrUnauthorized
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			return apperror.ErrTokenInvalid
		}

		claims, err := jwtService.ValidateAccessToken(parts[1])
		if err != nil {
			if err == infraAuth.ErrExpiredToken {
				return apperror.ErrTokenExpired
			}
			return apperror.ErrTokenInvalid
		}

		// IP binding check - token faqat yaratilgan IP dan ishlaydi
		if claims.IPAddress != "" && claims.IPAddress != c.IP() {
			return apperror.ErrTokenInvalid.WithMessage("Token cannot be used from a different IP address")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}
