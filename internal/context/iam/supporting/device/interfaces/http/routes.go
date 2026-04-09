package http

import "github.com/gofiber/fiber/v2"

// RegisterRoutes registers device management routes.
func (h *Handler) RegisterRoutes(api fiber.Router) {
	devices := api.Group("/devices")
	devices.Get("/", h.List)
	devices.Post("/trust", h.Trust)
	devices.Post("/untrust", h.Untrust)
	devices.Delete("/:id", h.Remove)
}
