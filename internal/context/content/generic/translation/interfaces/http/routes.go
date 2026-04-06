package http

import "github.com/gofiber/fiber/v2"

func (h *Handler) RegisterRoutes(group fiber.Router) {
	t := group.Group("/translations")
	t.Post("/", h.Create)
	t.Get("/", h.List)
	t.Get("/:id", h.GetByID)
	t.Patch("/:id", h.Update)
	t.Delete("/:id", h.Delete)
}
