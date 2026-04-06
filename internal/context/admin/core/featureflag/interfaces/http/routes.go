package http

import "github.com/gofiber/fiber/v2"

// RegisterRoutes registers feature flag routes under the given router group.
// Admin routes require ADMIN role (enforced by caller).
func (h *Handler) RegisterRoutes(admin fiber.Router, public fiber.Router) {
	// Admin CRUD
	flags := admin.Group("/flags")
	flags.Post("/", h.Create)
	flags.Get("/", h.List)
	flags.Get("/:id", h.GetByID)
	flags.Patch("/:id", h.Update)
	flags.Delete("/:id", h.Delete)

	// Public evaluation (authenticated users can check flags)
	public.Post("/flags/evaluate", h.Evaluate)
}
