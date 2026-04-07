package http

import (
	"net/http"

	kycApp "github.com/BakhodiribnYashinibnMansur/XBank/internal/application/kyc"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/domain/kyc"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/interfaces/http/dto"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type KYCHandler struct {
	service *kycApp.Service
}

func NewKYCHandler(service *kycApp.Service) *KYCHandler {
	return &KYCHandler{service: service}
}

// Submit godoc
// @Summary      Submit KYC verification
// @Tags         KYC
// @Accept       json
// @Produce      json
// @Param        body body dto.KYCSubmitRequest true "Document details"
// @Success      201 {object} dto.KYCResponse
// @Security     BearerAuth
// @Router       /kyc/submit [post]
func (h *KYCHandler) Submit(c *fiber.Ctx) error {
	var req dto.KYCSubmitRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.DocumentNumber == "" || req.FirstName == "" {
		return apperror.ErrMissingField.WithMessage("document_number and first_name are required")
	}

	userID := c.Locals("user_id").(string)
	v, err := h.service.Submit(c.Context(), userID, kyc.DocType(req.DocumentType), req.DocumentNumber, req.FirstName, req.LastName, req.DateOfBirth)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, toKYCResponse(v))
}

// Status godoc
// @Summary      Get KYC status
// @Tags         KYC
// @Produce      json
// @Success      200 {object} dto.KYCResponse
// @Security     BearerAuth
// @Router       /kyc/status [get]
func (h *KYCHandler) Status(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	v, err := h.service.GetStatus(c.Context(), userID)
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, toKYCResponse(v))
}

// Approve godoc
// @Summary      Approve KYC (admin)
// @Tags         KYC
// @Accept       json
// @Produce      json
// @Param        body body dto.KYCReviewRequest true "Verification ID"
// @Success      200 {object} map[string]string
// @Security     BearerAuth
// @Router       /admin/kyc/approve [post]
func (h *KYCHandler) Approve(c *fiber.Ctx) error {
	var req dto.KYCReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	reviewerID := c.Locals("user_id").(string)
	if err := h.service.Approve(c.Context(), req.VerificationID, reviewerID); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "KYC approved"})
}

// Reject godoc
// @Summary      Reject KYC (admin)
// @Tags         KYC
// @Accept       json
// @Produce      json
// @Param        body body dto.KYCReviewRequest true "Verification ID + reason"
// @Success      200 {object} map[string]string
// @Security     BearerAuth
// @Router       /admin/kyc/reject [post]
func (h *KYCHandler) Reject(c *fiber.Ctx) error {
	var req dto.KYCReviewRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	reviewerID := c.Locals("user_id").(string)
	if err := h.service.Reject(c.Context(), req.VerificationID, reviewerID, req.Reason); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "KYC rejected"})
}

// ListPending godoc
// @Summary      List pending KYC verifications (admin)
// @Tags         KYC
// @Produce      json
// @Param        page  query int false "Page" default(1)
// @Param        limit query int false "Limit" default(20)
// @Success      200 {object} dto.PaginatedResponse
// @Security     BearerAuth
// @Router       /admin/kyc/pending [get]
func (h *KYCHandler) ListPending(c *fiber.Ctx) error {
	pg := dto.ParsePagination(c)
	items, total, err := h.service.ListPending(c.Context(), pg.Limit, pg.Offset())
	if err != nil {
		return err
	}

	var data []dto.KYCResponse
	for _, v := range items {
		data = append(data, toKYCResponse(v))
	}
	if data == nil {
		data = []dto.KYCResponse{}
	}

	return apperror.Success(c, http.StatusOK, dto.PaginatedResponse{
		Data:       data,
		Pagination: dto.PaginationResponse{Page: pg.Page, Limit: pg.Limit, Total: total},
	})
}

func toKYCResponse(v *kyc.Verification) dto.KYCResponse {
	return dto.KYCResponse{
		ID: v.ID, UserID: v.UserID, DocumentType: string(v.DocumentType),
		FirstName: v.FirstName, LastName: v.LastName, Status: string(v.Status),
		RejectedReason: v.RejectedReason, CreatedAt: v.CreatedAt,
	}
}
