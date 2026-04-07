package http

import "github.com/gofiber/fiber/v2"

func (h *Handler) RegisterRoutes(group fiber.Router) {
	users := group.Group("/users")
	users.Get("/get", h.GetByID)
	users.Post("/change-password", h.ChangePassword)
	users.Get("/me/data-export", h.ExportData)
	users.Delete("/me/delete", h.DeleteAccount)
}
