package http

import "github.com/gofiber/fiber/v2"

func (h *Handler) RegisterRoutes(group fiber.Router) {
	audit := group.Group("/audit")
	audit.Get("/logs", h.ListAuditLogs)
	audit.Get("/endpoints", h.ListEndpointHistory)
}
