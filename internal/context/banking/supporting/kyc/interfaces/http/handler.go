package http

import (
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/kyc/application/command"
	kyc "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/banking/supporting/kyc/domain"
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

// Submit godoc
// @Summary      Submit KYC verification
// @Tags         KYC
// @Accept       json
// @Produce      json
// @Param        body body KYCSubmitRequest true "KYC document data"
// @Success      201 {object} KYCResponse
// @Failure      400 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /kyc/submit [post]
func (h *Handler) Submit(c *fiber.Ctx) error {
	var req KYCSubmitRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	userID := c.Locals("user_id").(string)
	v, err := h.service.Submit(
		c.Context(), userID,
		kyc.DocType(req.DocumentType), req.DocumentNumber,
		req.FirstName, req.LastName, req.DateOfBirth,
	)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, toKYCResponse(v))
}

// Status godoc
// @Summary      Get current user's KYC status
// @Tags         KYC
// @Produce      json
// @Success      200 {object} KYCResponse
// @Failure      404 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /kyc/status [get]
func (h *Handler) Status(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	v, err := h.service.GetStatus(c.Context(), userID)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, toKYCResponse(v))
}

// Approve godoc
// @Summary      Approve a KYC verification (admin)
// @Tags         Admin - KYC
// @Accept       json
// @Produce      json
// @Param        body body KYCReviewRequest true "Verification ID"
// @Success      200 {object} map[string]string
// @Failure      400 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/kyc/approve [post]
func (h *Handler) Approve(c *fiber.Ctx) error {
	var req KYCReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.VerificationID == "" {
		return apperror.ErrMissingField.WithMessage("verification_id is required")
	}

	reviewerID := c.Locals("user_id").(string)
	if err := h.service.Approve(c.Context(), req.VerificationID, reviewerID); err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "KYC approved"})
}

// Reject godoc
// @Summary      Reject a KYC verification (admin)
// @Tags         Admin - KYC
// @Accept       json
// @Produce      json
// @Param        body body KYCReviewRequest true "Verification ID and reason"
// @Success      200 {object} map[string]string
// @Failure      400 {object} apperror.ErrorResponse
// @Security     BearerAuth
// @Router       /admin/kyc/reject [post]
func (h *Handler) Reject(c *fiber.Ctx) error {
	var req KYCReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.VerificationID == "" {
		return apperror.ErrMissingField.WithMessage("verification_id is required")
	}
	if req.Reason == "" {
		return apperror.ErrMissingField.WithMessage("reason is required")
	}

	reviewerID := c.Locals("user_id").(string)
	if err := h.service.Reject(c.Context(), req.VerificationID, reviewerID, req.Reason); err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "KYC rejected"})
}

// ListPending godoc
// @Summary      List pending KYC verifications (admin, paginated)
// @Tags         Admin - KYC
// @Produce      json
// @Param        page  query int false "Page number" default(1)
// @Param        limit query int false "Items per page" default(20)
// @Success      200 {object} httpx.PaginatedResponse
// @Security     BearerAuth
// @Router       /admin/kyc/pending [get]
func (h *Handler) ListPending(c *fiber.Ctx) error {
	pg := httpx.ParsePagination(c)

	items, total, err := h.service.ListPending(c.Context(), pg.Limit, pg.Offset())
	if err != nil {
		return err
	}

	var data []KYCResponse
	for _, v := range items {
		data = append(data, toKYCResponse(v))
	}
	if data == nil {
		data = []KYCResponse{}
	}

	return apperror.Success(c, http.StatusOK, httpx.PaginatedResponse{
		Data:       data,
		Pagination: httpx.PaginationResponse{Page: pg.Page, Limit: pg.Limit, Total: total},
	})
}

func toKYCResponse(v *kyc.Verification) KYCResponse {
	return KYCResponse{
		ID:             v.ID,
		UserID:         v.UserID,
		DocumentType:   string(v.DocumentType),
		FirstName:      v.FirstName,
		LastName:       v.LastName,
		Status:         string(v.Status),
		RejectedReason: v.RejectedReason,
		CreatedAt:      v.CreatedAt,
	}
}
