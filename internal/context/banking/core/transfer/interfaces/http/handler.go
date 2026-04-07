package http

import (
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/application/command"
	transfer "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/httpx"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *command.Service
}

func NewHandler(service *command.Service) *Handler {
	return &Handler{service: service}
}

// Send godoc
// @Summary      Transfer funds between accounts
// @Tags         Transfers
// @Accept       json
// @Produce      json
// @Param        body body SendTransferRequest true "Transfer details"
// @Success      201 {object} TransferResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /transfers/send [post]
func (h *Handler) Send(c *fiber.Ctx) error {
	var req SendTransferRequest
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

	tr, err := h.service.Send(c.Context(), req.FromAccountID, req.ToAccountID, req.Amount, domain.Currency(req.Currency), req.Description)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, toTransferResponse(tr))
}

// GetByID godoc
// @Summary      Get transfer by ID
// @Tags         Transfers
// @Produce      json
// @Param        id query string true "Transfer ID"
// @Success      200 {object} TransferResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /transfers/get [get]
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return apperror.ErrMissingField.WithMessage("id query parameter is required")
	}

	tr, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, toTransferResponse(tr))
}

// ListByAccount godoc
// @Summary      List transfers for an account (paginated)
// @Tags         Transfers
// @Produce      json
// @Param        account_id query string true  "Account ID"
// @Param        page       query int    false "Page number" default(1)
// @Param        limit      query int    false "Items per page" default(20)
// @Success      200 {object} httpx.PaginatedResponse
// @Security     BearerAuth
// @Router       /transfers/list [get]
func (h *Handler) ListByAccount(c *fiber.Ctx) error {
	accountID := c.Query("account_id")
	if accountID == "" {
		return apperror.ErrMissingField.WithMessage("account_id query parameter is required")
	}

	pg := httpx.ParsePagination(c)
	transfers, total, err := h.service.ListByAccountID(c.Context(), accountID, pg.Limit, pg.Offset())
	if err != nil {
		return err
	}

	var data []TransferResponse
	for _, tr := range transfers {
		data = append(data, toTransferResponse(tr))
	}
	if data == nil {
		data = []TransferResponse{}
	}

	return apperror.Success(c, http.StatusOK, httpx.PaginatedResponse{
		Data:       data,
		Pagination: httpx.PaginationResponse{Page: pg.Page, Limit: pg.Limit, Total: total},
	})
}

// History godoc
// @Summary      Get transfer event history
// @Tags         Transfers
// @Produce      json
// @Param        id query string true "Transfer ID"
// @Success      200 {array} TransferEventResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /transfers/history [get]
func (h *Handler) History(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return apperror.ErrMissingField.WithMessage("id query parameter is required")
	}

	events, err := h.service.GetHistory(c.Context(), id)
	if err != nil {
		return err
	}

	var resp []TransferEventResponse
	for _, e := range events {
		resp = append(resp, TransferEventResponse{
			ID:        e.ID,
			Type:      string(e.Type),
			Data:      e.Data,
			Version:   e.Version,
			OccuredAt: e.OccurredAt,
		})
	}

	return apperror.Success(c, http.StatusOK, resp)
}

func toTransferResponse(t *transfer.Transfer) TransferResponse {
	return TransferResponse{
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
