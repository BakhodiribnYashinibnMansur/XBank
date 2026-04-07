package http

import "github.com/gofiber/fiber/v2"

func (h *Handler) RegisterRoutes(group fiber.Router) {
	ben := group.Group("/beneficiaries")
	ben.Post("/add", h.Add)
	ben.Get("/list", h.List)
	ben.Delete("/delete", h.Delete)
}
