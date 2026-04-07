package http

import (
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/exchange/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/exchange/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *command.Service
}

func NewHandler(service *command.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetRate(c *fiber.Ctx) error {
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

func (h *Handler) ListRates(c *fiber.Ctx) error {
	rates, err := h.service.ListAll(c.Context())
	if err != nil {
		return err
	}
	var resp []RateResponse
	for _, r := range rates {
		resp = append(resp, toRateResponse(r))
	}
	if resp == nil {
		resp = []RateResponse{}
	}
	return apperror.Success(c, http.StatusOK, resp)
}

func (h *Handler) Convert(c *fiber.Ctx) error {
	var req ConvertRequest
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

	return apperror.Success(c, http.StatusOK, ConvertResponse{
		From: req.From, To: req.To,
		OriginalAmount: req.Amount, ConvertedAmount: converted,
		RateUsed: rate.SellRate,
	})
}

func (h *Handler) UpsertRate(c *fiber.Ctx) error {
	var req UpsertRateRequest
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

func toRateResponse(r *domain.Rate) RateResponse {
	return RateResponse{
		ID: r.ID, FromCurrency: r.FromCurrency, ToCurrency: r.ToCurrency,
		BuyRate: r.BuyRate, SellRate: r.SellRate, UpdatedAt: r.UpdatedAt,
	}
}
