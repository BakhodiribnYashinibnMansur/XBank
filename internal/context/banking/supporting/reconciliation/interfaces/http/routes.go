package http

import "github.com/gofiber/fiber/v2"

func (h *Handler) RegisterRoutes(group fiber.Router) {
	recon := group.Group("/admin/reconciliation")
	recon.Get("/check", h.CheckAccount)
	recon.Get("/check-all", h.CheckAllByUser)
}
