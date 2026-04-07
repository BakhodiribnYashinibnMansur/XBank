package http

import "github.com/gofiber/fiber/v2"

func (h *Handler) RegisterRoutes(group fiber.Router) {
	challenge := group.Group("/auth/challenge")
	challenge.Post("/request", h.Request)
	challenge.Post("/verify", h.Verify)
}
