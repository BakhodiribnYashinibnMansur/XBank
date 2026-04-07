package http

import "github.com/gofiber/fiber/v2"

// RegisterRoutes registers customer-facing KYC routes.
func (h *Handler) RegisterRoutes(group fiber.Router) {
	kyc := group.Group("/kyc")
	kyc.Post("/submit", h.Submit)
	kyc.Get("/status", h.Status)
}

// RegisterAdminRoutes registers admin KYC routes.
func (h *Handler) RegisterAdminRoutes(admin fiber.Router) {
	kyc := admin.Group("/kyc")
	kyc.Post("/approve", h.Approve)
	kyc.Post("/reject", h.Reject)
	kyc.Get("/pending", h.ListPending)
}
