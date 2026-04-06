package http

import "github.com/gofiber/fiber/v2"

func (h *Handler) RegisterRoutes(admin fiber.Router) {
	admin.Get("/statistics/overview", h.Overview)
}
