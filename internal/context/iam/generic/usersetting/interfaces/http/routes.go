package http

import "github.com/gofiber/fiber/v2"

// RegisterRoutes registers user settings routes.
func (h *Handler) RegisterRoutes(api fiber.Router) {
	settings := api.Group("/settings")
	settings.Put("/", h.Upsert)
	settings.Get("/", h.List)
	settings.Delete("/:id", h.Delete)
}
