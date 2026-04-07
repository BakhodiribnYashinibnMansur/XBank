package http

import (
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/beneficiary/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/beneficiary/domain"
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

func (h *Handler) Add(c *fiber.Ctx) error {
	var req AddBeneficiaryRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	userID := c.Locals("user_id").(string)
	b, err := h.service.Add(c.Context(), userID, req.Name, req.AccountNumber, req.BankName, req.BankCode, req.Currency, domain.Type(req.Type))
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, toBenResponse(b))
}

func (h *Handler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	pg := httpx.ParsePagination(c)

	items, total, err := h.service.ListByUserID(c.Context(), userID, pg.Limit, pg.Offset())
	if err != nil {
		return err
	}

	var data []BeneficiaryResponse
	for _, b := range items {
		data = append(data, toBenResponse(b))
	}
	if data == nil {
		data = []BeneficiaryResponse{}
	}

	return apperror.Success(c, http.StatusOK, httpx.PaginatedResponse{
		Data:       data,
		Pagination: httpx.PaginationResponse{Page: pg.Page, Limit: pg.Limit, Total: total},
	})
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return apperror.ErrMissingField.WithMessage("id is required")
	}
	if err := h.service.Delete(c.Context(), id); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "Beneficiary deleted"})
}

func toBenResponse(b *domain.Beneficiary) BeneficiaryResponse {
	return BeneficiaryResponse{
		ID: b.ID, Name: b.Name, AccountNumber: b.AccountNumber,
		BankName: b.BankName, BankCode: b.BankCode, Currency: b.Currency,
		Type: string(b.BenType), Verified: b.Verified, CreatedAt: b.CreatedAt,
	}
}
