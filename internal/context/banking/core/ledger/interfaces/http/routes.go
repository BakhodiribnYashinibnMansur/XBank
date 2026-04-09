package http

import "github.com/gofiber/fiber/v2"

// RegisterRoutes registers ledger read-only routes.
func (h *Handler) RegisterRoutes(api fiber.Router) {
	ledger := api.Group("/ledger")
	ledger.Get("/", h.ListByAccount)
}
