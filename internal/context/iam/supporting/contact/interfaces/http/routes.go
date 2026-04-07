package http

import "github.com/gofiber/fiber/v2"

func (h *Handler) RegisterRoutes(group fiber.Router) {
	contacts := group.Group("/contacts")
	contacts.Post("/add", h.Add)
	contacts.Get("/list", h.List)
	contacts.Delete("/delete", h.Delete)
}
