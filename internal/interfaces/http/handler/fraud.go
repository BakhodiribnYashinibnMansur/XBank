package handler

import (
	fraudApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/fraud"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/fraud"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/dto"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type FraudHandler struct {
	service *fraudApp.Service
}

func NewFraudHandler(service *fraudApp.Service) *FraudHandler {
	return &FraudHandler{service: service}
}

// ListFlagged godoc
// @Summary      List flagged/blocked transactions (admin)
// @Tags         Fraud
// @Produce      json
// @Param        page  query int false "Page" default(1)
// @Param        limit query int false "Limit" default(20)
// @Success      200 {object} dto.PaginatedResponse
// @Security     BearerAuth
// @Router       /admin/fraud/flagged [get]
func (h *FraudHandler) ListFlagged(c *fiber.Ctx) error {
	pg := dto.ParsePagination(c)
	items, total, err := h.service.ListFlagged(c.Context(), pg.Limit, pg.Offset())
	if err != nil {
		return err
	}

	var data []dto.FraudCheckResponse
	for _, ch := range items {
		data = append(data, toFraudResponse(ch))
	}
	if data == nil {
		data = []dto.FraudCheckResponse{}
	}

	return c.JSON(dto.PaginatedResponse{
		Data:       data,
		Pagination: dto.PaginationResponse{Page: pg.Page, Limit: pg.Limit, Total: total},
	})
}

// GetByTransfer godoc
// @Summary      Get fraud check for a transfer
// @Tags         Fraud
// @Produce      json
// @Param        transfer_id query string true "Transfer ID"
// @Success      200 {object} dto.FraudCheckResponse
// @Security     BearerAuth
// @Router       /admin/fraud/check [get]
func (h *FraudHandler) GetByTransfer(c *fiber.Ctx) error {
	transferID := c.Query("transfer_id")
	if transferID == "" {
		return apperror.ErrMissingField.WithMessage("transfer_id is required")
	}

	check, err := h.service.GetByTransferID(c.Context(), transferID)
	if err != nil {
		return err
	}
	return c.JSON(toFraudResponse(check))
}

func toFraudResponse(ch *fraud.Check) dto.FraudCheckResponse {
	return dto.FraudCheckResponse{
		ID: ch.ID, TransferID: ch.TransferID, UserID: ch.UserID,
		RiskScore: ch.RiskScore, RiskLevel: string(ch.RiskLevel),
		Action: string(ch.Action), Flags: ch.Flags,
		ReviewedBy: ch.ReviewedBy, ReviewComment: ch.ReviewComment,
		CreatedAt: ch.CreatedAt,
	}
}
