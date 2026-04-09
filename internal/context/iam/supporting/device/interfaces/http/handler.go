package http

import (
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/supporting/device/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *command.Service
}

func NewHandler(service *command.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return apperror.ErrUnauthorized
	}

	devices, err := h.service.List(c.Context(), userID)
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, devices)
}

type trustRequest struct {
	DeviceHash string `json:"device_hash"`
}

func (h *Handler) Trust(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return apperror.ErrUnauthorized
	}

	var req trustRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	if err := h.service.Trust(c.Context(), userID, req.DeviceHash); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"trusted": true})
}

func (h *Handler) Untrust(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	if userID == "" {
		return apperror.ErrUnauthorized
	}

	var req trustRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	if err := h.service.Untrust(c.Context(), userID, req.DeviceHash); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"trusted": false})
}

func (h *Handler) Remove(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.service.Remove(c.Context(), id); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"deleted": true})
}
