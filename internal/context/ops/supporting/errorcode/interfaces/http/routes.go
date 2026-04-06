package http

import "github.com/gofiber/fiber/v2"

func (h *Handler) RegisterRoutes(admin fiber.Router) {
	ec := admin.Group("/errorcodes")
	ec.Post("/", h.Create)
	ec.Get("/", h.List)
	ec.Get("/lookup/:code", h.Lookup)
	ec.Patch("/:id", h.Update)
	ec.Delete("/:id", h.Delete)
}
