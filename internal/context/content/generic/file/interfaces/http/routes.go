package http

import "github.com/gofiber/fiber/v2"

// RegisterRoutes registers file management routes.
func (h *Handler) RegisterRoutes(api fiber.Router) {
	files := api.Group("/files")
	files.Get("/", h.List)
	files.Get("/:id", h.GetByID)
	files.Delete("/:id", h.Delete)
}
