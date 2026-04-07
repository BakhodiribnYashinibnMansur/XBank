package http

import (
	"net/http"

	command "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/application/command"
	account "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/core/account/domain"
	kernelDomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
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

// Create godoc
// @Summary      Create a new bank account
// @Tags         Accounts
// @Accept       json
// @Produce      json
// @Param        body body CreateAccountRequest true "Currency (UZS, USD, EUR)"
// @Success      201 {object} AccountResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /accounts/create [post]
func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateAccountRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	currency := kernelDomain.Currency(req.Currency)
	if currency != kernelDomain.UZS && currency != kernelDomain.USD && currency != kernelDomain.EUR {
		return apperror.ErrValidation.WithMessage("Currency must be UZS, USD or EUR")
	}

	userID := c.Locals("user_id").(string)
	acc, err := h.service.CreateAccount(c.Context(), userID, currency)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, toAccountResponse(acc))
}

// GetByID godoc
// @Summary      Get account by ID
// @Tags         Accounts
// @Produce      json
// @Param        id query string true "Account ID"
// @Success      200 {object} AccountResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /accounts/get [get]
func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return apperror.ErrMissingField.WithMessage("id query parameter is required")
	}

	acc, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, toAccountResponse(acc))
}

// List godoc
// @Summary      List user accounts (paginated)
// @Tags         Accounts
// @Produce      json
// @Param        page  query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(20)
// @Success      200 {object} httpx.PaginatedResponse
// @Security     BearerAuth
// @Router       /accounts/list [get]
func (h *Handler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	pg := httpx.ParsePagination(c)

	accounts, total, err := h.service.ListByUserID(c.Context(), userID, pg.Limit, pg.Offset())
	if err != nil {
		return err
	}

	var data []AccountResponse
	for _, acc := range accounts {
		data = append(data, toAccountResponse(acc))
	}
	if data == nil {
		data = []AccountResponse{}
	}

	return apperror.Success(c, http.StatusOK, httpx.PaginatedResponse{
		Data:       data,
		Pagination: httpx.PaginationResponse{Page: pg.Page, Limit: pg.Limit, Total: total},
	})
}

// Deposit godoc
// @Summary      Deposit funds
// @Tags         Accounts
// @Accept       json
// @Produce      json
// @Param        body body DepositRequest true "Account ID and amount (in minor units)"
// @Success      200 {object} AccountResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /accounts/deposit [post]
func (h *Handler) Deposit(c *fiber.Ctx) error {
	var req DepositRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.AccountID == "" {
		return apperror.ErrMissingField.WithMessage("account_id is required")
	}
	if req.Amount <= 0 {
		return apperror.ErrValidation.WithMessage("Amount must be greater than 0")
	}

	acc, err := h.service.Deposit(c.Context(), req.AccountID, req.Amount)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, toAccountResponse(acc))
}

// Withdraw godoc
// @Summary      Withdraw funds
// @Tags         Accounts
// @Accept       json
// @Produce      json
// @Param        body body WithdrawRequest true "Account ID and amount (in minor units)"
// @Success      200 {object} AccountResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /accounts/withdraw [post]
func (h *Handler) Withdraw(c *fiber.Ctx) error {
	var req WithdrawRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.AccountID == "" {
		return apperror.ErrMissingField.WithMessage("account_id is required")
	}
	if req.Amount <= 0 {
		return apperror.ErrValidation.WithMessage("Amount must be greater than 0")
	}

	acc, err := h.service.Withdraw(c.Context(), req.AccountID, req.Amount)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, toAccountResponse(acc))
}

// Close godoc
// @Summary      Close an account (balance must be 0)
// @Tags         Accounts
// @Accept       json
// @Produce      json
// @Param        body body CloseAccountRequest true "Account ID"
// @Success      200 {object} map[string]string
// @Failure      400 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /accounts/close [post]
func (h *Handler) Close(c *fiber.Ctx) error {
	var req CloseAccountRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.AccountID == "" {
		return apperror.ErrMissingField.WithMessage("account_id is required")
	}

	if err := h.service.CloseAccount(c.Context(), req.AccountID); err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "Account closed"})
}

// History godoc
// @Summary      Get account event history
// @Tags         Accounts
// @Produce      json
// @Param        id query string true "Account ID"
// @Success      200 {array} AccountEventResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /accounts/history [get]
func (h *Handler) History(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return apperror.ErrMissingField.WithMessage("id query parameter is required")
	}

	events, err := h.service.GetHistory(c.Context(), id)
	if err != nil {
		return err
	}

	var resp []AccountEventResponse
	for _, e := range events {
		resp = append(resp, AccountEventResponse{
			ID:        e.ID,
			Type:      string(e.Type),
			Data:      e.Data,
			Version:   e.Version,
			OccuredAt: e.OccurredAt,
		})
	}

	return apperror.Success(c, http.StatusOK, resp)
}

func toAccountResponse(a *account.Account) AccountResponse {
	return AccountResponse{
		ID:            a.ID,
		AccountNumber: a.AccountNumber,
		Balance:       a.Balance.Amount,
		Currency:      string(a.Balance.Currency),
		Status:        string(a.Status),
		CreatedAt:     a.CreatedAt,
	}
}
