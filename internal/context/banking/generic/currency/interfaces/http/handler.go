package http

import (
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/generic/currency/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/generic/currency/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

// Handler handles HTTP requests for currency management.
type Handler struct {
	service *command.Service
}

// NewHandler creates a new currency HTTP handler.
func NewHandler(service *command.Service) *Handler {
	return &Handler{service: service}
}

// Create godoc
// @Summary      Create a new currency
// @Tags         Currencies
// @Accept       json
// @Produce      json
// @Param        body body CreateCurrencyRequest true "Currency details"
// @Success      201 {object} CurrencyResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /currencies/create [post]
func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateCurrencyRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.Code == "" {
		return apperror.ErrMissingField.WithMessage("code is required")
	}
	if req.Name == "" {
		return apperror.ErrMissingField.WithMessage("name is required")
	}

	cur, err := h.service.Create(c.Context(), req.Code, req.Name, req.Symbol, req.DecimalPlaces)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, toCurrencyResponse(cur))
}

// GetByCode godoc
// @Summary      Get currency by code
// @Tags         Currencies
// @Produce      json
// @Param        code query string true "Currency code (e.g. USD)"
// @Success      200 {object} CurrencyResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /currencies/get [get]
func (h *Handler) GetByCode(c *fiber.Ctx) error {
	code := c.Query("code")
	if code == "" {
		return apperror.ErrMissingField.WithMessage("code query parameter is required")
	}

	cur, err := h.service.GetByCode(c.Context(), code)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, toCurrencyResponse(cur))
}

// List godoc
// @Summary      List all currencies
// @Tags         Currencies
// @Produce      json
// @Success      200 {array} CurrencyResponse
// @Security     BearerAuth
// @Router       /currencies/list [get]
func (h *Handler) List(c *fiber.Ctx) error {
	currencies, err := h.service.ListAll(c.Context())
	if err != nil {
		return err
	}

	var resp []CurrencyResponse
	for _, cur := range currencies {
		resp = append(resp, toCurrencyResponse(cur))
	}
	if resp == nil {
		resp = []CurrencyResponse{}
	}

	return apperror.Success(c, http.StatusOK, resp)
}

// Update godoc
// @Summary      Update a currency
// @Tags         Currencies
// @Accept       json
// @Produce      json
// @Param        body body UpdateCurrencyRequest true "Updated currency details"
// @Success      200 {object} CurrencyResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /currencies/update [put]
func (h *Handler) Update(c *fiber.Ctx) error {
	var req UpdateCurrencyRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.ID == "" {
		return apperror.ErrMissingField.WithMessage("id is required")
	}

	cur, err := h.service.Update(c.Context(), req.ID, req.Name, req.Symbol, req.DecimalPlaces)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, toCurrencyResponse(cur))
}

// ToggleStatus godoc
// @Summary      Activate or deactivate a currency
// @Tags         Currencies
// @Accept       json
// @Produce      json
// @Param        body body ToggleStatusRequest true "Currency ID and desired status"
// @Success      200 {object} CurrencyResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /currencies/toggle-status [post]
func (h *Handler) ToggleStatus(c *fiber.Ctx) error {
	var req ToggleStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.ID == "" {
		return apperror.ErrMissingField.WithMessage("id is required")
	}

	cur, err := h.service.ToggleStatus(c.Context(), req.ID, req.Active)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, toCurrencyResponse(cur))
}

func toCurrencyResponse(c *domain.Currency) CurrencyResponse {
	return CurrencyResponse{
		ID:            c.ID,
		Code:          c.Code,
		Name:          c.Name,
		Symbol:        c.Symbol,
		DecimalPlaces: c.DecimalPlaces,
		Status:        string(c.Status),
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}
