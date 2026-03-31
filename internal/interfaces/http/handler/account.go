package handler

import (
	"net/http"

	accountApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/account"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/account"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/shared"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/dto"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type AccountHandler struct {
	service *accountApp.Service
}

func NewAccountHandler(service *accountApp.Service) *AccountHandler {
	return &AccountHandler{service: service}
}

// Create godoc
// @Summary      Create a new bank account
// @Tags         Accounts
// @Accept       json
// @Produce      json
// @Param        body body dto.CreateAccountRequest true "Currency (UZS, USD, EUR)"
// @Success      201 {object} dto.AccountResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /accounts/create [post]
func (h *AccountHandler) Create(c *fiber.Ctx) error {
	var req dto.CreateAccountRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	currency := shared.Currency(req.Currency)
	if currency != shared.UZS && currency != shared.USD && currency != shared.EUR {
		return apperror.ErrValidation.WithMessage("Currency must be UZS, USD or EUR")
	}

	userID := c.Locals("user_id").(string)
	acc, err := h.service.CreateAccount(c.Context(), userID, currency)
	if err != nil {
		return err
	}

	return c.Status(http.StatusCreated).JSON(toAccountResponse(acc))
}

// GetByID godoc
// @Summary      Get account by ID
// @Tags         Accounts
// @Produce      json
// @Param        id query string true "Account ID"
// @Success      200 {object} dto.AccountResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /accounts/get [get]
func (h *AccountHandler) GetByID(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return apperror.ErrMissingField.WithMessage("id query parameter is required")
	}

	acc, err := h.service.GetByID(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(toAccountResponse(acc))
}

// List godoc
// @Summary      List user accounts (paginated)
// @Tags         Accounts
// @Produce      json
// @Param        page  query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(20)
// @Success      200 {object} dto.PaginatedResponse
// @Security     BearerAuth
// @Router       /accounts/list [get]
func (h *AccountHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	pg := dto.ParsePagination(c)

	accounts, total, err := h.service.ListByUserID(c.Context(), userID, pg.Limit, pg.Offset())
	if err != nil {
		return err
	}

	var data []dto.AccountResponse
	for _, acc := range accounts {
		data = append(data, toAccountResponse(acc))
	}
	if data == nil {
		data = []dto.AccountResponse{}
	}

	return c.JSON(dto.PaginatedResponse{
		Data:       data,
		Pagination: dto.PaginationResponse{Page: pg.Page, Limit: pg.Limit, Total: total},
	})
}

// Deposit godoc
// @Summary      Deposit funds
// @Tags         Accounts
// @Accept       json
// @Produce      json
// @Param        body body dto.DepositRequest true "Account ID and amount (in minor units)"
// @Success      200 {object} dto.AccountResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /accounts/deposit [post]
func (h *AccountHandler) Deposit(c *fiber.Ctx) error {
	var req dto.DepositRequest
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

	return c.JSON(toAccountResponse(acc))
}

// Withdraw godoc
// @Summary      Withdraw funds
// @Tags         Accounts
// @Accept       json
// @Produce      json
// @Param        body body dto.WithdrawRequest true "Account ID and amount (in minor units)"
// @Success      200 {object} dto.AccountResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /accounts/withdraw [post]
func (h *AccountHandler) Withdraw(c *fiber.Ctx) error {
	var req dto.WithdrawRequest
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

	return c.JSON(toAccountResponse(acc))
}

// Close godoc
// @Summary      Close an account (balance must be 0)
// @Tags         Accounts
// @Accept       json
// @Produce      json
// @Param        body body dto.CloseAccountRequest true "Account ID"
// @Success      200 {object} map[string]string
// @Failure      400 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /accounts/close [post]
func (h *AccountHandler) Close(c *fiber.Ctx) error {
	var req dto.CloseAccountRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.AccountID == "" {
		return apperror.ErrMissingField.WithMessage("account_id is required")
	}

	if err := h.service.CloseAccount(c.Context(), req.AccountID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "Account closed"})
}

func toAccountResponse(a *account.Account) dto.AccountResponse {
	return dto.AccountResponse{
		ID:            a.ID,
		AccountNumber: a.AccountNumber,
		Balance:       a.Balance.Amount,
		Currency:      string(a.Balance.Currency),
		Status:        string(a.Status),
		CreatedAt:     a.CreatedAt,
	}
}
