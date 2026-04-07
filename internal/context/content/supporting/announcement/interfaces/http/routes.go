package http

import "github.com/gofiber/fiber/v2"

// RegisterAdminRoutes registers admin CRUD routes for announcements.
func (h *Handler) RegisterAdminRoutes(admin fiber.Router) {
	a := admin.Group("/announcements")
	a.Post("/", h.Create)
	a.Get("/", h.List)
	a.Get("/:id", h.GetByID)
	a.Post("/:id/publish", h.Publish)
	a.Delete("/:id", h.Delete)
}
