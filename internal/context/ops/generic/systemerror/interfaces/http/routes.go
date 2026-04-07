package http

import "github.com/gofiber/fiber/v2"

func (h *Handler) RegisterRoutes(admin fiber.Router) {
	e := admin.Group("/errors")
	e.Get("/", h.List)
	e.Post("/:id/resolve", h.Resolve)
}
