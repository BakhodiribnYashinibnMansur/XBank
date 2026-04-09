package http

import "github.com/gofiber/fiber/v2"

// RegisterAdminRoutes registers admin routes for data export management.
func (h *Handler) RegisterAdminRoutes(admin fiber.Router) {
	exports := admin.Group("/exports")
	exports.Post("/", h.Request)
	exports.Get("/", h.List)
	exports.Get("/:id", h.GetByID)
}
