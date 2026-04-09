package http

import "github.com/gofiber/fiber/v2"

// RegisterRoutes registers admin dashboard routes.
func (h *Handler) RegisterRoutes(group fiber.Router) {
	dashboard := group.Group("/admin/dashboard")
	dashboard.Get("/overview", h.Overview)
	dashboard.Get("/period", h.PeriodStats)
}
