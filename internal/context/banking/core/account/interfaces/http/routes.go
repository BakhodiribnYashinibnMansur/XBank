package http

import "github.com/gofiber/fiber/v2"

func (h *Handler) RegisterRoutes(group fiber.Router) {
	accounts := group.Group("/accounts")
	accounts.Post("/create", h.Create)
	accounts.Get("/get", h.GetByID)
	accounts.Get("/list", h.List)
	accounts.Post("/deposit", h.Deposit)
	accounts.Post("/withdraw", h.Withdraw)
	accounts.Post("/close", h.Close)
	accounts.Get("/history", h.History)
}
