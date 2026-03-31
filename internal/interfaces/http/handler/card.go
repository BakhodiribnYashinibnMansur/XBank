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
// @Router       /cards/issue [post]
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

	return c.Status(http.StatusCreated).JSON(toCardResponse(card))
}

// Activate godoc
// @Summary      Activate a card by setting PIN
// @Tags         Cards
// @Accept       json
// @Produce      json
// @Param        body body dto.ActivateCardRequest true "Card ID and 4-digit PIN"
// @Success      200 {object} dto.CardResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /cards/activate [post]
func (h *CardHandler) Activate(c *fiber.Ctx) error {
	var req dto.ActivateCardRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.CardID == "" || req.PIN == "" {
		return apperror.ErrMissingField.WithMessage("card_id and pin are required")
	}

	card, err := h.service.Activate(c.Context(), req.CardID, req.PIN)
	if err != nil {
		return err
	}

	return c.JSON(toCardResponse(card))
}

// VerifyPIN godoc
// @Summary      Verify card PIN
// @Tags         Cards
// @Accept       json
// @Produce      json
// @Param        body body dto.VerifyPINRequest true "Card ID and PIN"
// @Success      200 {object} map[string]string
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      403 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /cards/verify-pin [post]
func (h *CardHandler) VerifyPIN(c *fiber.Ctx) error {
	var req dto.VerifyPINRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.CardID == "" || req.PIN == "" {
		return apperror.ErrMissingField.WithMessage("card_id and pin are required")
	}

	if err := h.service.VerifyPIN(c.Context(), req.CardID, req.PIN); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "PIN verified"})
}

// ChangePIN godoc
// @Summary      Change card PIN
// @Tags         Cards
// @Accept       json
// @Produce      json
// @Param        body body dto.ChangePINRequest true "Card ID, old PIN, new PIN"
// @Success      200 {object} map[string]string
// @Failure      400 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /cards/change-pin [post]
func (h *CardHandler) ChangePIN(c *fiber.Ctx) error {
	var req dto.ChangePINRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.CardID == "" || req.OldPIN == "" || req.NewPIN == "" {
		return apperror.ErrMissingField.WithMessage("card_id, old_pin and new_pin are required")
	}

	if err := h.service.ChangePIN(c.Context(), req.CardID, req.OldPIN, req.NewPIN); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "PIN changed"})
}

// Block godoc
// @Summary      Block a card
// @Tags         Cards
// @Accept       json
// @Produce      json
// @Param        body body dto.CardActionRequest true "Card ID"
// @Success      200 {object} map[string]string
// @Security     BearerAuth
// @Router       /cards/block [post]
func (h *CardHandler) Block(c *fiber.Ctx) error {
	var req dto.CardActionRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.CardID == "" {
		return apperror.ErrMissingField.WithMessage("card_id is required")
	}

	if err := h.service.Block(c.Context(), req.CardID); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "Card blocked"})
}

// Unblock godoc
// @Summary      Unblock a card
// @Tags         Cards
// @Accept       json
// @Produce      json
// @Param        body body dto.CardActionRequest true "Card ID"
// @Success      200 {object} map[string]string
// @Security     BearerAuth
// @Router       /cards/unblock [post]
func (h *CardHandler) Unblock(c *fiber.Ctx) error {
	var req dto.CardActionRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.CardID == "" {
		return apperror.ErrMissingField.WithMessage("card_id is required")
	}

	if err := h.service.Unblock(c.Context(), req.CardID); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "Card unblocked"})
}

// GetByID godoc
// @Summary      Get card by ID
// @Tags         Cards
// @Produce      json
// @Param        id query string true "Card ID"
// @Success      200 {object} dto.CardResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /cards/get [get]
func (h *CardHandler) GetByID(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return apperror.ErrMissingField.WithMessage("id query parameter is required")
	}

	card, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(toCardResponse(card))
}

// List godoc
// @Summary      List cards for an account (paginated)
// @Tags         Cards
// @Produce      json
// @Param        account_id query string true  "Account ID"
// @Param        page       query int    false "Page number" default(1)
// @Param        limit      query int    false "Items per page" default(20)
// @Success      200 {object} dto.PaginatedResponse
// @Security     BearerAuth
// @Router       /cards/list [get]
func (h *CardHandler) List(c *fiber.Ctx) error {
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

	return c.JSON(dto.PaginatedResponse{
		Data:       data,
		Pagination: dto.PaginationResponse{Page: pg.Page, Limit: pg.Limit, Total: total},
	})
}

func toCardResponse(c *domainCard.Card) dto.CardResponse {
	return dto.CardResponse{
		ID:          c.ID,
		AccountID:   c.AccountID,
		MaskedPAN:   c.MaskedPAN, // Never return full PAN!
		ExpiryMonth: c.ExpiryMonth,
		ExpiryYear:  c.ExpiryYear,
		CardType:    string(c.CardType),
		Status:      string(c.Status),
		CreatedAt:   c.CreatedAt,
	}
}
