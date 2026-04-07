package http

import "github.com/gofiber/fiber/v2"

func (h *Handler) RegisterRoutes(group fiber.Router) {
	cards := group.Group("/cards")

	// Core card operations
	cards.Post("/", h.Issue)
	cards.Get("/", h.ByAccount)
	cards.Get("/:id", h.ByID)
	cards.Post("/:id/activate", h.Activate)
	cards.Post("/:id/verify-pin", h.VerifyPIN)
	cards.Put("/:id/pin", h.ChangePIN)
	cards.Post("/:id/block", h.Block)
	cards.Post("/:id/unblock", h.Unblock)
}

func (h *ExtendedHandler) RegisterRoutes(group fiber.Router) {
	cards := group.Group("/cards")

	// Tokenization
	cards.Post("/:id/tokenize", h.Tokenize)
	cards.Get("/:id/tokens", h.ListTokens)
	cards.Post("/tokens/revoke", h.RevokeToken)

	// Holds
	cards.Post("/holds", h.CreateHold)
	cards.Post("/holds/:id/capture", h.CaptureHold)
	cards.Post("/holds/:id/release", h.ReleaseHold)
	cards.Get("/:id/holds", h.ListHolds)

	// 3D Secure / EMV
	cards.Post("/:id/3ds/enroll", h.Enroll3DS)
	cards.Post("/:id/emv", h.SetEMV)
}
