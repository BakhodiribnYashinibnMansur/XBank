package http

import "github.com/gofiber/fiber/v2"

// RegisterRoutes registers template management routes.
func (h *Handler) RegisterRoutes(group fiber.Router) {
	templates := group.Group("/templates")
	templates.Post("/create", h.Create)
	templates.Get("/get", h.GetByID)
	templates.Get("/resolve", h.Resolve)
	templates.Get("/list", h.List)
	templates.Put("/update-body", h.UpdateBody)
	templates.Post("/activate", h.Activate)
	templates.Post("/archive", h.Archive)
}
