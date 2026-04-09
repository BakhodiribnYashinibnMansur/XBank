package middleware

import (
	"strings"

	infraAuth "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/security/jwt"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
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
			logger.Log.Warn("auth_failure: missing authorization header",
				zap.String("ip", c.IP()),
				zap.String("path", c.Path()),
				zap.String("user_agent", c.Get("User-Agent")),
			)
			return apperror.ErrUnauthorized
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Log.Warn("auth_failure: malformed bearer token",
				zap.String("ip", c.IP()),
				zap.String("path", c.Path()),
				zap.String("user_agent", c.Get("User-Agent")),
			)
			return apperror.ErrTokenInvalid
		}

		claims, err := jwtService.ValidateAccessToken(parts[1])
		if err != nil {
			reason := "invalid_token"
			if err == infraAuth.ErrExpiredToken {
				reason = "expired_token"
			}
			logger.Log.Warn("auth_failure: token validation failed",
				zap.String("reason", reason),
				zap.String("ip", c.IP()),
				zap.String("path", c.Path()),
				zap.String("user_agent", c.Get("User-Agent")),
			)
			if err == infraAuth.ErrExpiredToken {
				return apperror.ErrTokenExpired
			}
			return apperror.ErrTokenInvalid
		}

		// IP binding check - token faqat yaratilgan IP dan ishlaydi
		if claims.IPAddress != "" && claims.IPAddress != c.IP() {
			logger.Log.Warn("auth_failure: IP binding mismatch",
				zap.String("user_id", claims.UserID),
				zap.String("token_ip", claims.IPAddress),
				zap.String("request_ip", c.IP()),
				zap.String("path", c.Path()),
			)
			return apperror.ErrTokenInvalid.WithMessage("Token cannot be used from a different IP address")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}
