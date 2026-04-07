package http

import "github.com/gofiber/fiber/v2"

func (h *Handler) RegisterRoutes(group fiber.Router) {
	n := group.Group("/notifications")
	n.Get("/", h.List)
	n.Get("/:id", h.GetByID)
	n.Post("/:id/read", h.MarkAsRead)
	n.Delete("/:id", h.Delete)
}
