package http

import "github.com/gofiber/fiber/v2"

func (h *Handler) RegisterRoutes(group fiber.Router) {
	fraud := group.Group("/admin/fraud")
	fraud.Get("/flagged", h.ListFlagged)
	fraud.Get("/check", h.GetByTransfer)
}
