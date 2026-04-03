package handler

import (
	"net/http"

	exchApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/exchange"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/exchange"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/dto"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type ExchangeHandler struct {
	service *exchApp.Service
}

func NewExchangeHandler(service *exchApp.Service) *ExchangeHandler {
	return &ExchangeHandler{service: service}
}

// GetRate godoc
// @Summary      Get exchange rate
// @Tags         Exchange
// @Produce      json
// @Param        from query string true "From currency (e.g. USD)"
// @Param        to   query string true "To currency (e.g. UZS)"
// @Success      200 {object} dto.RateResponse
// @Router       /currencies/rate [get]
func (h *ExchangeHandler) GetRate(c *fiber.Ctx) error {
	from, to := c.Query("from"), c.Query("to")
	if from == "" || to == "" {
		return apperror.ErrMissingField.WithMessage("from and to query parameters are required")
	}

	rate, err := h.service.GetRate(c.Context(), from, to)
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, toRateResponse(rate))
}

// ListRates godoc
// @Summary      List all exchange rates
// @Tags         Exchange
// @Produce      json
// @Success      200 {array} dto.RateResponse
// @Router       /currencies/rates [get]
func (h *ExchangeHandler) ListRates(c *fiber.Ctx) error {
	rates, err := h.service.ListAll(c.Context())
	if err != nil {
		return err
	}
	var resp []dto.RateResponse
	for _, r := range rates {
		resp = append(resp, toRateResponse(r))
	}
	if resp == nil {
		resp = []dto.RateResponse{}
	}
	return apperror.Success(c, http.StatusOK, resp)
}

// Convert godoc
// @Summary      Convert currency
// @Tags         Exchange
// @Accept       json
// @Produce      json
// @Param        body body dto.ConvertRequest true "Conversion details"
// @Success      200 {object} dto.ConvertResponse
// @Security     BearerAuth
// @Router       /currencies/convert [post]
func (h *ExchangeHandler) Convert(c *fiber.Ctx) error {
	var req dto.ConvertRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.From == "" || req.To == "" || req.Amount <= 0 {
		return apperror.ErrMissingField.WithMessage("from, to and amount are required")
	}

	converted, rate, err := h.service.Convert(c.Context(), req.From, req.To, req.Amount)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, dto.ConvertResponse{
		From: req.From, To: req.To,
		OriginalAmount: req.Amount, ConvertedAmount: converted,
		RateUsed: rate.SellRate,
	})
}

// UpsertRate godoc
// @Summary      Set/update exchange rate (admin only)
// @Tags         Exchange
// @Accept       json
// @Produce      json
// @Param        body body dto.UpsertRateRequest true "Rate details"
// @Success      200 {object} dto.RateResponse
// @Security     BearerAuth
// @Router       /currencies/rate [post]
func (h *ExchangeHandler) UpsertRate(c *fiber.Ctx) error {
	var req dto.UpsertRateRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.From == "" || req.To == "" || req.BuyRate <= 0 || req.SellRate <= 0 {
		return apperror.ErrValidation.WithMessage("from, to, buy_rate and sell_rate are required")
	}

	rate, err := h.service.UpsertRate(c.Context(), req.From, req.To, req.BuyRate, req.SellRate)
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, toRateResponse(rate))
}

func toRateResponse(r *exchange.Rate) dto.RateResponse {
	return dto.RateResponse{
		ID: r.ID, FromCurrency: r.FromCurrency, ToCurrency: r.ToCurrency,
		BuyRate: r.BuyRate, SellRate: r.SellRate, UpdatedAt: r.UpdatedAt,
	}
}
