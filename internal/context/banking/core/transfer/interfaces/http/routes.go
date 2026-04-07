package http

import "github.com/gofiber/fiber/v2"

// RegisterRoutes registers all Transfer BC HTTP routes.
func RegisterRoutes(router fiber.Router, h *Handler, sh *ScheduledHandler) {
	transfers := router.Group("/transfers")

	// Transfer endpoints
	transfers.Post("/send", h.Send)
	transfers.Get("/get", h.GetByID)
	transfers.Get("/list", h.ListByAccount)
	transfers.Get("/history", h.History)

	// Scheduled transfer endpoints
	scheduled := transfers.Group("/scheduled")
	scheduled.Post("/", sh.Schedule)
	scheduled.Post("/cancel", sh.Cancel)
	scheduled.Get("/get", sh.GetByID)
	scheduled.Get("/list", sh.List)
}
