package http

import (
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/fraud/application/command"
	fraud "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/fraud/domain"
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

// ListFlagged godoc
// @Summary      List flagged/blocked fraud checks (paginated)
// @Tags         Admin - Fraud
// @Produce      json
// @Param        page  query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(20)
// @Success      200 {object} httpx.PaginatedResponse
// @Security     BearerAuth
// @Router       /admin/fraud/flagged [get]
func (h *Handler) ListFlagged(c *fiber.Ctx) error {
	pg := httpx.ParsePagination(c)

	items, total, err := h.service.ListFlagged(c.Context(), pg.Limit, pg.Offset())
	if err != nil {
		return err
	}

	var data []FraudCheckResponse
	for _, ch := range items {
		data = append(data, toFraudCheckResponse(ch))
	}
	if data == nil {
		data = []FraudCheckResponse{}
	}

	return apperror.Success(c, http.StatusOK, httpx.PaginatedResponse{
		Data:       data,
		Pagination: httpx.PaginationResponse{Page: pg.Page, Limit: pg.Limit, Total: total},
	})
}

// GetByTransfer godoc
// @Summary      Get fraud check by transfer ID
// @Tags         Admin - Fraud
// @Produce      json
// @Param        transfer_id query string true "Transfer ID"
// @Success      200 {object} FraudCheckResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/fraud/check [get]
func (h *Handler) GetByTransfer(c *fiber.Ctx) error {
	transferID := c.Query("transfer_id")
	if transferID == "" {
		return apperror.ErrMissingField.WithMessage("transfer_id is required")
	}

	ch, err := h.service.GetByTransferID(c.Context(), transferID)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, toFraudCheckResponse(ch))
}

func toFraudCheckResponse(ch *fraud.Check) FraudCheckResponse {
	return FraudCheckResponse{
		ID:            ch.ID,
		TransferID:    ch.TransferID,
		UserID:        ch.UserID,
		RiskScore:     ch.RiskScore,
		RiskLevel:     string(ch.RiskLevel),
		Action:        string(ch.Action),
		Flags:         ch.Flags,
		ReviewedBy:    ch.ReviewedBy,
		ReviewComment: ch.ReviewComment,
		CreatedAt:     ch.CreatedAt,
	}
}
