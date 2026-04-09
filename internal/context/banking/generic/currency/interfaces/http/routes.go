package http

import "github.com/gofiber/fiber/v2"

// RegisterRoutes registers currency management routes.
func (h *Handler) RegisterRoutes(group fiber.Router) {
	currencies := group.Group("/currencies")
	currencies.Post("/create", h.Create)
	currencies.Get("/get", h.GetByCode)
	currencies.Get("/list", h.List)
	currencies.Put("/update", h.Update)
	currencies.Post("/toggle-status", h.ToggleStatus)
}
