package http

import (
	"net/http"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/application/command"
	transfer "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/transfer/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/httpx"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type ScheduledHandler struct {
	service *command.ScheduledService
}

func NewScheduledHandler(service *command.ScheduledService) *ScheduledHandler {
	return &ScheduledHandler{service: service}
}

// Schedule godoc
// @Summary      Schedule a future transfer
// @Tags         Scheduled Transfers
// @Accept       json
// @Produce      json
// @Param        body body ScheduleTransferRequest true "Scheduled transfer details"
// @Success      201 {object} ScheduledTransferResponse
// @Security     BearerAuth
// @Router       /transfers/scheduled [post]
func (h *ScheduledHandler) Schedule(c *fiber.Ctx) error {
	var req ScheduleTransferRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	if req.FromAccountID == "" || req.ToAccountID == "" || req.Currency == "" {
		return apperror.ErrMissingField.WithMessage("from_account_id, to_account_id, and currency are required")
	}
	if req.Amount <= 0 {
		return apperror.ErrValidation.WithMessage("Amount must be greater than 0")
	}

	executeAt, err := time.Parse(time.RFC3339, req.ExecuteAt)
	if err != nil {
		return apperror.ErrValidation.WithMessage("execute_at must be in RFC3339 format (e.g. 2026-04-03T10:00:00Z)")
	}

	userID := c.Locals("user_id").(string)

	st, err := h.service.Schedule(c.Context(), userID, req.FromAccountID, req.ToAccountID, req.Amount, domain.Currency(req.Currency), req.Description, executeAt)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, toScheduledResponse(st))
}

// Cancel godoc
// @Summary      Cancel a scheduled transfer
// @Tags         Scheduled Transfers
// @Param        id query string true "Scheduled transfer ID"
// @Success      200 {object} map[string]string
// @Security     BearerAuth
// @Router       /transfers/scheduled/cancel [post]
func (h *ScheduledHandler) Cancel(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return apperror.ErrMissingField.WithMessage("id query parameter is required")
	}

	userID := c.Locals("user_id").(string)

	if err := h.service.Cancel(c.Context(), id, userID); err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, fiber.Map{"status": "cancelled"})
}

// GetByID godoc
// @Summary      Get a scheduled transfer by ID
// @Tags         Scheduled Transfers
// @Param        id query string true "Scheduled transfer ID"
// @Success      200 {object} ScheduledTransferResponse
// @Security     BearerAuth
// @Router       /transfers/scheduled/get [get]
func (h *ScheduledHandler) GetByID(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return apperror.ErrMissingField.WithMessage("id query parameter is required")
	}

	st, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, toScheduledResponse(st))
}

// List godoc
// @Summary      List scheduled transfers for the current user
// @Tags         Scheduled Transfers
// @Param        page  query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(20)
// @Success      200 {object} httpx.PaginatedResponse
// @Security     BearerAuth
// @Router       /transfers/scheduled/list [get]
func (h *ScheduledHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	pg := httpx.ParsePagination(c)

	items, total, err := h.service.ListByUserID(c.Context(), userID, pg.Limit, pg.Offset())
	if err != nil {
		return err
	}

	var data []ScheduledTransferResponse
	for _, st := range items {
		data = append(data, toScheduledResponse(st))
	}
	if data == nil {
		data = []ScheduledTransferResponse{}
	}

	return apperror.Success(c, http.StatusOK, httpx.PaginatedResponse{
		Data:       data,
		Pagination: httpx.PaginationResponse{Page: pg.Page, Limit: pg.Limit, Total: total},
	})
}

func toScheduledResponse(st *transfer.ScheduledTransfer) ScheduledTransferResponse {
	return ScheduledTransferResponse{
		ID:            st.ID,
		FromAccountID: st.FromAccountID,
		ToAccountID:   st.ToAccountID,
		Amount:        st.Amount.Amount,
		Currency:      string(st.Amount.Currency),
		Description:   st.Description,
		Status:        string(st.Status),
		ExecuteAt:     st.ExecuteAt.Format(time.RFC3339),
		TransferID:    st.TransferID,
		FailureReason: st.FailureReason,
		CreatedAt:     st.CreatedAt.Format(time.RFC3339),
	}
}
