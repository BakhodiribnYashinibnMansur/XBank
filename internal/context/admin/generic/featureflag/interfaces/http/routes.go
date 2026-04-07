package http

import "github.com/gofiber/fiber/v2"

// RegisterAdminRoutes registers admin CRUD routes for feature flags.
func (h *Handler) RegisterAdminRoutes(admin fiber.Router) {
	flags := admin.Group("/flags")
	flags.Post("/", h.Create)
	flags.Get("/", h.List)
	flags.Get("/:id", h.GetByID)
	flags.Patch("/:id", h.Update)
	flags.Delete("/:id", h.Delete)
}
