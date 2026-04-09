package http

import "github.com/gofiber/fiber/v2"

// RegisterRoutes registers health check routes.
func (h *Handler) RegisterRoutes(group fiber.Router) {
	health := group.Group("/health")
	health.Post("/check", h.Check)
	health.Get("/latest", h.Latest)
	health.Get("/history", h.History)
}
