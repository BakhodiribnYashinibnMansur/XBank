package handler

import (
	"net/http"

	transferApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/transfer"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/transfer"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/dto"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type TransferHandler struct {
	service *transferApp.Service
}

func NewTransferHandler(service *transferApp.Service) *TransferHandler {
	return &TransferHandler{service: service}
}

// Send godoc
// @Summary      Transfer funds between accounts
// @Tags         Transfers
// @Accept       json
// @Produce      json
// @Param        body body dto.SendTransferRequest true "Transfer details"
// @Success      201 {object} dto.TransferResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /transfers/send [post]
func (h *TransferHandler) Send(c *fiber.Ctx) error {
	var req dto.SendTransferRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	if req.FromAccountID == "" || req.ToAccountID == "" {
		return apperror.ErrMissingField.WithMessage("from_account_id and to_account_id are required")
	}
	if req.Amount <= 0 {
		return apperror.ErrValidation.WithMessage("Amount must be greater than 0")
	}
	if req.Currency == "" {
		return apperror.ErrMissingField.WithMessage("currency is required")
	}

	tr, err := h.service.Send(c.Context(), req.FromAccountID, req.ToAccountID, req.Amount, shared.Currency(req.Currency), req.Description)
	if err != nil {
		return err
	}

	return c.Status(http.StatusCreated).JSON(toTransferResponse(tr))
}

// GetByID godoc
// @Summary      Get transfer by ID
// @Tags         Transfers
// @Produce      json
// @Param        id query string true "Transfer ID"
// @Success      200 {object} dto.TransferResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /transfers/get [get]
func (h *TransferHandler) GetByID(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return apperror.ErrMissingField.WithMessage("id query parameter is required")
	}

	tr, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(toTransferResponse(tr))
}

// ListByAccount godoc
// @Summary      List transfers for an account (paginated)
// @Tags         Transfers
// @Produce      json
// @Param        account_id query string true  "Account ID"
// @Param        page       query int    false "Page number" default(1)
// @Param        limit      query int    false "Items per page" default(20)
// @Success      200 {object} dto.PaginatedResponse
// @Security     BearerAuth
// @Router       /transfers/list [get]
func (h *TransferHandler) ListByAccount(c *fiber.Ctx) error {
	accountID := c.Query("account_id")
	if accountID == "" {
		return apperror.ErrMissingField.WithMessage("account_id query parameter is required")
	}

	pg := dto.ParsePagination(c)
	transfers, total, err := h.service.ListByAccountID(c.Context(), accountID, pg.Limit, pg.Offset())
	if err != nil {
		return err
	}

	var data []dto.TransferResponse
	for _, tr := range transfers {
		data = append(data, toTransferResponse(tr))
	}
	if data == nil {
		data = []dto.TransferResponse{}
	}

	return c.JSON(dto.PaginatedResponse{
		Data:       data,
		Pagination: dto.PaginationResponse{Page: pg.Page, Limit: pg.Limit, Total: total},
	})
}

func toTransferResponse(t *transfer.Transfer) dto.TransferResponse {
	return dto.TransferResponse{
		ID:            t.ID,
		FromAccountID: t.FromAccountID,
		ToAccountID:   t.ToAccountID,
		Amount:        t.Amount.Amount,
		Currency:      string(t.Amount.Currency),
		Status:        string(t.Status),
		Description:   t.Description,
		FailureReason: t.FailureReason,
		CreatedAt:     t.CreatedAt,
	}
}
