package handler

import (
	"net/http"

	cardApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/card"
	domainCard "github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/card"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/dto"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type CardHandler struct {
	service *cardApp.Service
}

func NewCardHandler(service *cardApp.Service) *CardHandler {
	return &CardHandler{service: service}
}

// Issue godoc
// @Summary      Issue a new card
// @Tags         Cards
// @Accept       json
// @Produce      json
// @Param        body body dto.IssueCardRequest true "Account ID and card type"
// @Success      201 {object} dto.CardResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /cards [post]
func (h *CardHandler) Issue(c *fiber.Ctx) error {
	var req dto.IssueCardRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.AccountID == "" {
		return apperror.ErrMissingField.WithMessage("account_id is required")
	}

	cardType := domainCard.Type(req.CardType)
	if cardType != domainCard.TypeDebit && cardType != domainCard.TypeVirtual {
		return apperror.ErrValidation.WithMessage("card_type must be DEBIT or VIRTUAL")
	}

	card, err := h.service.IssueCard(c.Context(), req.AccountID, cardType)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, toCardResponse(card))
}

// Activate godoc
// @Summary      Activate a card by setting PIN
// @Tags         Cards
// @Accept       json
// @Produce      json
// @Param        id  path string true "Card ID"
// @Param        body body dto.ActivateCardRequest true "4-digit PIN"
// @Success      200 {object} dto.CardResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /cards/{id}/activate [post]
func (h *CardHandler) Activate(c *fiber.Ctx) error {
	cardID := c.Params("id")

	var req dto.ActivateCardRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.PIN == "" {
		return apperror.ErrMissingField.WithMessage("pin is required")
	}

	card, err := h.service.Activate(c.Context(), cardID, req.PIN)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, toCardResponse(card))
}

// VerifyPIN godoc
// @Summary      Verify card PIN
// @Tags         Cards
// @Accept       json
// @Produce      json
// @Param        id  path string true "Card ID"
// @Param        body body dto.VerifyPINRequest true "PIN"
// @Success      200 {object} map[string]string
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      403 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /cards/{id}/verify-pin [post]
func (h *CardHandler) VerifyPIN(c *fiber.Ctx) error {
	cardID := c.Params("id")

	var req dto.VerifyPINRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.PIN == "" {
		return apperror.ErrMissingField.WithMessage("pin is required")
	}

	if err := h.service.VerifyPIN(c.Context(), cardID, req.PIN); err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "PIN verified"})
}

// ChangePIN godoc
// @Summary      Change card PIN
// @Tags         Cards
// @Accept       json
// @Produce      json
// @Param        id  path string true "Card ID"
// @Param        body body dto.ChangePINRequest true "Old PIN and new PIN"
// @Success      200 {object} map[string]string
// @Failure      400 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /cards/{id}/pin [put]
func (h *CardHandler) ChangePIN(c *fiber.Ctx) error {
	cardID := c.Params("id")

	var req dto.ChangePINRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.OldPIN == "" || req.NewPIN == "" {
		return apperror.ErrMissingField.WithMessage("old_pin and new_pin are required")
	}

	if err := h.service.ChangePIN(c.Context(), cardID, req.OldPIN, req.NewPIN); err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "PIN changed"})
}

// Block godoc
// @Summary      Block a card
// @Tags         Cards
// @Produce      json
// @Param        id path string true "Card ID"
// @Success      200 {object} map[string]string
// @Security     BearerAuth
// @Router       /cards/{id}/block [post]
func (h *CardHandler) Block(c *fiber.Ctx) error {
	cardID := c.Params("id")

	if err := h.service.Block(c.Context(), cardID); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "Card blocked"})
}

// Unblock godoc
// @Summary      Unblock a card
// @Tags         Cards
// @Produce      json
// @Param        id path string true "Card ID"
// @Success      200 {object} map[string]string
// @Security     BearerAuth
// @Router       /cards/{id}/unblock [post]
func (h *CardHandler) Unblock(c *fiber.Ctx) error {
	cardID := c.Params("id")

	if err := h.service.Unblock(c.Context(), cardID); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "Card unblocked"})
}

// ByID godoc
// @Summary      Card details by ID
// @Tags         Cards
// @Produce      json
// @Param        id path string true "Card ID"
// @Success      200 {object} dto.CardResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /cards/{id} [get]
func (h *CardHandler) ByID(c *fiber.Ctx) error {
	cardID := c.Params("id")

	card, err := h.service.GetByID(c.Context(), cardID)
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, toCardResponse(card))
}

// ByAccount godoc
// @Summary      Cards for an account (paginated)
// @Tags         Cards
// @Produce      json
// @Param        account_id query string true  "Account ID"
// @Param        page       query int    false "Page number" default(1)
// @Param        limit      query int    false "Items per page" default(20)
// @Success      200 {object} dto.PaginatedResponse
// @Security     BearerAuth
// @Router       /cards [get]
func (h *CardHandler) ByAccount(c *fiber.Ctx) error {
	accountID := c.Query("account_id")
	if accountID == "" {
		return apperror.ErrMissingField.WithMessage("account_id query parameter is required")
	}

	pg := dto.ParsePagination(c)
	cards, total, err := h.service.ListByAccountID(c.Context(), accountID, pg.Limit, pg.Offset())
	if err != nil {
		return err
	}

	var data []dto.CardResponse
	for _, card := range cards {
		data = append(data, toCardResponse(card))
	}
	if data == nil {
		data = []dto.CardResponse{}
	}

	return apperror.Success(c, http.StatusOK, dto.PaginatedResponse{
		Data:       data,
		Pagination: dto.PaginationResponse{Page: pg.Page, Limit: pg.Limit, Total: total},
	})
}

func toCardResponse(c *domainCard.Card) dto.CardResponse {
	return dto.CardResponse{
		ID:          c.ID,
		AccountID:   c.AccountID,
		MaskedPAN:   c.MaskedPAN,
		ExpiryMonth: c.ExpiryMonth,
		ExpiryYear:  c.ExpiryYear,
		CardType:    string(c.CardType),
		Status:      string(c.Status),
		CreatedAt:   c.CreatedAt,
	}
}
