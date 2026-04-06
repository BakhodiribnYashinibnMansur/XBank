package http

import "github.com/gofiber/fiber/v2"

func (h *Handler) RegisterRoutes(admin fiber.Router, public fiber.Router) {
	// Admin CRUD
	a := admin.Group("/announcements")
	a.Post("/", h.Create)
	a.Get("/", h.List)
	a.Get("/:id", h.GetByID)
	a.Post("/:id/publish", h.Publish)
	a.Delete("/:id", h.Delete)

	// Public: active announcements
	public.Get("/announcements/active", h.ListActive)
}
