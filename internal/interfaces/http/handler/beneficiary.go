package handler

import (
	"net/http"

	benApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/beneficiary"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/beneficiary"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/dto"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type BeneficiaryHandler struct {
	service *benApp.Service
}

func NewBeneficiaryHandler(service *benApp.Service) *BeneficiaryHandler {
	return &BeneficiaryHandler{service: service}
}

// Add godoc
// @Summary      Add a beneficiary
// @Tags         Beneficiaries
// @Accept       json
// @Produce      json
// @Param        body body dto.AddBeneficiaryRequest true "Beneficiary details"
// @Success      201 {object} dto.BeneficiaryResponse
// @Security     BearerAuth
// @Router       /beneficiaries/add [post]
func (h *BeneficiaryHandler) Add(c *fiber.Ctx) error {
	var req dto.AddBeneficiaryRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	userID := c.Locals("user_id").(string)
	b, err := h.service.Add(c.Context(), userID, req.Name, req.AccountNumber, req.BankName, req.BankCode, req.Currency, beneficiary.Type(req.Type))
	if err != nil {
		return err
	}

	return c.Status(http.StatusCreated).JSON(toBenResponse(b))
}

// List godoc
// @Summary      List beneficiaries (paginated)
// @Tags         Beneficiaries
// @Produce      json
// @Param        page  query int false "Page" default(1)
// @Param        limit query int false "Limit" default(20)
// @Success      200 {object} dto.PaginatedResponse
// @Security     BearerAuth
// @Router       /beneficiaries/list [get]
func (h *BeneficiaryHandler) List(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	pg := dto.ParsePagination(c)

	items, total, err := h.service.ListByUserID(c.Context(), userID, pg.Limit, pg.Offset())
	if err != nil {
		return err
	}

	var data []dto.BeneficiaryResponse
	for _, b := range items {
		data = append(data, toBenResponse(b))
	}
	if data == nil {
		data = []dto.BeneficiaryResponse{}
	}

	return c.JSON(dto.PaginatedResponse{
		Data:       data,
		Pagination: dto.PaginationResponse{Page: pg.Page, Limit: pg.Limit, Total: total},
	})
}

// Delete godoc
// @Summary      Delete a beneficiary
// @Tags         Beneficiaries
// @Produce      json
// @Param        id query string true "Beneficiary ID"
// @Success      200 {object} map[string]string
// @Security     BearerAuth
// @Router       /beneficiaries/delete [delete]
func (h *BeneficiaryHandler) Delete(c *fiber.Ctx) error {
	id := c.Query("id")
	if id == "" {
		return apperror.ErrMissingField.WithMessage("id is required")
	}
	if err := h.service.Delete(c.Context(), id); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"message": "Beneficiary deleted"})
}

func toBenResponse(b *beneficiary.Beneficiary) dto.BeneficiaryResponse {
	return dto.BeneficiaryResponse{
		ID: b.ID, Name: b.Name, AccountNumber: b.AccountNumber,
		BankName: b.BankName, BankCode: b.BankCode, Currency: b.Currency,
		Type: string(b.BenType), Verified: b.Verified, CreatedAt: b.CreatedAt,
	}
}
