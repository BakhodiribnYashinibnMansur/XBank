package http

import "github.com/gofiber/fiber/v2"

// RegisterRoutes registers site setting routes under the given router group.
// All routes require ADMIN role (enforced by caller).
func (h *Handler) RegisterRoutes(group fiber.Router) {
	settings := group.Group("/settings")
	settings.Put("/", h.Upsert)
	settings.Get("/", h.List)
	settings.Get("/get", h.GetByID)
	settings.Delete("/", h.Delete)
}
