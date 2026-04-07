package http

import "github.com/gofiber/fiber/v2"

func (h *Handler) RegisterRoutes(group fiber.Router) {
	currencies := group.Group("/currencies")
	currencies.Get("/rate", h.GetRate)
	currencies.Get("/rates", h.ListRates)
	currencies.Post("/convert", h.Convert)
	currencies.Post("/rate", h.UpsertRate)
}
