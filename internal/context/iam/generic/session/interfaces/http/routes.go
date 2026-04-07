package http

import "github.com/gofiber/fiber/v2"

func (h *Handler) RegisterRoutes(group fiber.Router) {
	auth := group.Group("/auth")
	auth.Post("/login", h.Login)
	auth.Post("/refresh", h.Refresh)
	auth.Post("/logout", h.Logout)
	auth.Post("/logout-all", h.LogoutAll)

	totp := auth.Group("/totp")
	totp.Post("/verify", h.TOTPVerifyLogin)
	totp.Post("/setup", h.TOTPSetup)
	totp.Post("/confirm", h.TOTPConfirmSetup)
	totp.Post("/disable", h.TOTPDisable)
}
