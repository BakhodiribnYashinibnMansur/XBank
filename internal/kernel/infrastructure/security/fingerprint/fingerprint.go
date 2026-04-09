package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Attributes holds the raw signals used to compute a device fingerprint.
type Attributes struct {
	UserAgent      string
	AcceptLanguage string
	IPAddress      string
	ScreenInfo     string // optional, from X-Screen-Info header
	Timezone       string // optional, from X-Timezone header
}

// Generator computes deterministic device fingerprint hashes from request attributes.
type Generator struct {
	includeIP bool
}

// NewGenerator creates a Generator. If includeIP is true, the client IP
// becomes part of the fingerprint (stricter but changes on network switch).
func NewGenerator(includeIP bool) *Generator {
	return &Generator{includeIP: includeIP}
}

// FromRequest extracts attributes from a Fiber request context and computes
// the fingerprint.
func (g *Generator) FromRequest(c *fiber.Ctx) string {
	attrs := Attributes{
		UserAgent:      c.Get("User-Agent"),
		AcceptLanguage: c.Get("Accept-Language"),
		IPAddress:      c.IP(),
		ScreenInfo:     c.Get("X-Screen-Info"),
		Timezone:       c.Get("X-Timezone"),
	}
	return g.FromAttributes(attrs)
}

// FromAttributes computes a SHA-256 fingerprint from the given attributes.
func (g *Generator) FromAttributes(attrs Attributes) string {
	var parts []string
	parts = append(parts, attrs.UserAgent)
	parts = append(parts, attrs.AcceptLanguage)
	if g.includeIP {
		parts = append(parts, attrs.IPAddress)
	}
	parts = append(parts, attrs.ScreenInfo)
	parts = append(parts, attrs.Timezone)

	data := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
