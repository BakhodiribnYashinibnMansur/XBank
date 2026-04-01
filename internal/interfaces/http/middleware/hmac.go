package middleware

import (
	"strconv"

	infraCrypto "github.com/BakhodiribnYashinibnMansur/XBank/internal/infrastructure/crypto"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

// HMACMiddleware - verifies request body integrity via HMAC-SHA256.
//
// Client must send two headers:
//   - X-Signature:           hex-encoded HMAC-SHA256(secret, timestamp + "." + body)
//   - X-Signature-Timestamp: Unix timestamp (seconds) when the request was signed
//
// The middleware:
//  1. Extracts headers
//  2. Validates timestamp freshness (anti-replay)
//  3. Recomputes HMAC and compares using constant-time comparison
//  4. Rejects tampered or expired requests
//
// Only applies to mutating methods (POST, PUT, PATCH, DELETE).
// GET/HEAD/OPTIONS pass through without verification.
func HMACMiddleware(signer *infraCrypto.HMACSigner) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip non-mutating methods
		method := c.Method()
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			return c.Next()
		}

		// If signer is nil (HMAC not configured), skip
		if signer == nil {
			return c.Next()
		}

		signature := c.Get("X-Signature")
		timestampStr := c.Get("X-Signature-Timestamp")

		if signature == "" || timestampStr == "" {
			return apperror.ErrHMACMissing
		}

		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			return apperror.ErrHMACTimestampInvalid
		}

		body := c.Body()

		if err := signer.Verify(signature, timestamp, body); err != nil {
			// Distinguish between expired and invalid
			if timestamp != 0 {
				// Try to determine if it's a timestamp issue
				// by checking if recomputed signature matches
				expected := signer.Sign(timestamp, body)
				if expected == signature {
					// Signature is correct but timestamp expired
					return apperror.ErrHMACTimestampExpired
				}
			}
			return apperror.ErrHMACSignatureInvalid
		}

		return c.Next()
	}
}
