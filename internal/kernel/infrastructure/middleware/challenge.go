package middleware

import (
	"context"

	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

// ChallengeValidator - interface for token validation (avoids circular import)
type ChallengeValidator interface {
	ValidateToken(ctx context.Context, token, userID string) error
}

// RequireChallenge - middleware that enforces step-up authentication.
//
// Use on sensitive endpoints (high-value transfers, card issuance, etc.)
// Client must include X-Challenge-Token header with a verified challenge token.
//
// Flow:
//  1. Extract X-Challenge-Token from request header
//  2. Validate token against Redis (fast) or DB (fallback)
//  3. Ensure token belongs to the authenticated user
//  4. Allow request if valid, reject otherwise
func RequireChallenge(validator ChallengeValidator) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// If validator is nil (challenge not configured), skip
		if validator == nil {
			return c.Next()
		}

		token := c.Get("X-Challenge-Token")
		if token == "" {
			return apperror.ErrChallengeTokenMissing
		}

		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return apperror.ErrUnauthorized
		}

		if err := validator.ValidateToken(c.Context(), token, userID); err != nil {
			return err
		}

		return c.Next()
	}
}
