package http

import (
	"net/http"
	"time"

	cardApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/card/application/command"
	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/card/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

// ExtendedHandler - handles tokenization, holds, 3DS, EMV
type ExtendedHandler struct {
	cardService  *cardApp.Service
	tokenService *cardApp.TokenService
	holdService  *cardApp.HoldService
}

func NewExtendedHandler(cardService *cardApp.Service, tokenService *cardApp.TokenService, holdService *cardApp.HoldService) *ExtendedHandler {
	return &ExtendedHandler{
		cardService:  cardService,
		tokenService: tokenService,
		holdService:  holdService,
	}
}

// --- Tokenization ---

// Tokenize godoc
// @Summary      Create a token for a card's PAN
// @Tags         Cards
// @Param        id path string true "Card ID"
// @Success      201 {object} TokenResponse
// @Security     BearerAuth
// @Router       /cards/{id}/tokenize [post]
func (h *ExtendedHandler) Tokenize(c *fiber.Ctx) error {
	cardID := c.Params("id")

	t, err := h.tokenService.Tokenize(c.Context(), cardID)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, TokenResponse{
		Token:    t.Token,
		CardID:   t.CardID,
		LastFour: t.LastFour,
		IsActive: t.IsActive,
	})
}

// ListTokens godoc
// @Summary      List active tokens for a card
// @Tags         Cards
// @Param        id path string true "Card ID"
// @Success      200 {array} TokenResponse
// @Security     BearerAuth
// @Router       /cards/{id}/tokens [get]
func (h *ExtendedHandler) ListTokens(c *fiber.Ctx) error {
	cardID := c.Params("id")

	tokens, err := h.tokenService.ListTokens(c.Context(), cardID)
	if err != nil {
		return err
	}

	var resp []TokenResponse
	for _, t := range tokens {
		resp = append(resp, TokenResponse{
			Token:    t.Token,
			CardID:   t.CardID,
			LastFour: t.LastFour,
			IsActive: t.IsActive,
		})
	}
	if resp == nil {
		resp = []TokenResponse{}
	}

	return apperror.Success(c, http.StatusOK, resp)
}

// RevokeToken godoc
// @Summary      Revoke a card token
// @Tags         Cards
// @Param        token query string true "Token to revoke"
// @Success      200 {object} map[string]string
// @Security     BearerAuth
// @Router       /cards/tokens/revoke [post]
func (h *ExtendedHandler) RevokeToken(c *fiber.Ctx) error {
	token := c.Query("token")
	if token == "" {
		return apperror.ErrMissingField.WithMessage("token query parameter is required")
	}

	if err := h.tokenService.RevokeToken(c.Context(), token); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "Token revoked"})
}

// --- Hold / Capture / Release ---

// CreateHold godoc
// @Summary      Create an authorization hold
// @Tags         Cards
// @Accept       json
// @Produce      json
// @Param        body body HoldRequest true "Hold details"
// @Success      201 {object} HoldResponse
// @Security     BearerAuth
// @Router       /cards/holds [post]
func (h *ExtendedHandler) CreateHold(c *fiber.Ctx) error {
	var req HoldRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.CardID == "" || req.AccountID == "" || req.Amount <= 0 || req.Currency == "" {
		return apperror.ErrMissingField.WithMessage("card_id, account_id, amount, and currency are required")
	}

	hold, err := h.holdService.Hold(c.Context(), req.CardID, req.AccountID, req.Merchant, req.Amount, req.Currency)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, toHoldResponse(hold))
}

// CaptureHold godoc
// @Summary      Capture (settle) an authorization hold
// @Tags         Cards
// @Param        id path string true "Hold ID"
// @Success      200 {object} HoldResponse
// @Security     BearerAuth
// @Router       /cards/holds/{id}/capture [post]
func (h *ExtendedHandler) CaptureHold(c *fiber.Ctx) error {
	holdID := c.Params("id")

	hold, err := h.holdService.Capture(c.Context(), holdID)
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, toHoldResponse(hold))
}

// ReleaseHold godoc
// @Summary      Release (cancel) an authorization hold
// @Tags         Cards
// @Param        id path string true "Hold ID"
// @Success      200 {object} HoldResponse
// @Security     BearerAuth
// @Router       /cards/holds/{id}/release [post]
func (h *ExtendedHandler) ReleaseHold(c *fiber.Ctx) error {
	holdID := c.Params("id")

	hold, err := h.holdService.Release(c.Context(), holdID)
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, toHoldResponse(hold))
}

// ListHolds godoc
// @Summary      List holds for a card
// @Tags         Cards
// @Param        id path string true "Card ID"
// @Success      200 {array} HoldResponse
// @Security     BearerAuth
// @Router       /cards/{id}/holds [get]
func (h *ExtendedHandler) ListHolds(c *fiber.Ctx) error {
	cardID := c.Params("id")

	holds, err := h.holdService.ListByCard(c.Context(), cardID)
	if err != nil {
		return err
	}

	var resp []HoldResponse
	for _, hold := range holds {
		resp = append(resp, toHoldResponse(hold))
	}
	if resp == nil {
		resp = []HoldResponse{}
	}
	return apperror.Success(c, http.StatusOK, resp)
}

// --- 3D Secure / EMV ---

// Enroll3DS godoc
// @Summary      Enroll a card in 3D Secure
// @Tags         Cards
// @Accept       json
// @Param        id   path string true "Card ID"
// @Param        body body Enroll3DSRequest true "3DS version"
// @Success      200 {object} CardResponse
// @Security     BearerAuth
// @Router       /cards/{id}/3ds/enroll [post]
func (h *ExtendedHandler) Enroll3DS(c *fiber.Ctx) error {
	cardID := c.Params("id")
	var req Enroll3DSRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.Version == "" {
		req.Version = "2.2"
	}

	updatedCard, err := h.cardService.Enroll3DS(c.Context(), cardID, req.Version)
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, toCardResponse(updatedCard))
}

// SetEMV godoc
// @Summary      Set EMV Application Identifier for a card
// @Tags         Cards
// @Accept       json
// @Param        id   path string true "Card ID"
// @Param        body body SetEMVRequest true "EMV AID"
// @Success      200 {object} CardResponse
// @Security     BearerAuth
// @Router       /cards/{id}/emv [post]
func (h *ExtendedHandler) SetEMV(c *fiber.Ctx) error {
	cardID := c.Params("id")
	var req SetEMVRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.AID == "" {
		return apperror.ErrMissingField.WithMessage("aid is required")
	}

	updatedCard, err := h.cardService.SetEMVAID(c.Context(), cardID, req.AID)
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, toCardResponse(updatedCard))
}

func toHoldResponse(h *domain.Hold) HoldResponse {
	return HoldResponse{
		ID:        h.ID,
		CardID:    h.CardID,
		AccountID: h.AccountID,
		Merchant:  h.Merchant,
		Amount:    h.Amount,
		Currency:  h.Currency,
		Status:    string(h.Status),
		HeldAt:    h.HeldAt.Format(time.RFC3339),
		ExpiresAt: h.ExpiresAt.Format(time.RFC3339),
	}
}
