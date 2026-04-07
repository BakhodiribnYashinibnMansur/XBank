package http

import (
	"net/http"
	"strconv"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/generic/featureflag/application"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/generic/featureflag/application/command"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/generic/featureflag/application/query"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/admin/generic/featureflag/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

// Handler handles HTTP requests for feature flags.
type Handler struct {
	create   *command.CreateHandler
	update   *command.UpdateHandler
	delete   *command.DeleteHandler
	get      *query.GetHandler
	list     *query.ListHandler
	evaluate *query.EvaluateHandler
}

func NewHandler(
	create *command.CreateHandler,
	update *command.UpdateHandler,
	del *command.DeleteHandler,
	get *query.GetHandler,
	list *query.ListHandler,
	evaluate *query.EvaluateHandler,
) *Handler {
	return &Handler{create: create, update: update, delete: del, get: get, list: list, evaluate: evaluate}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateFlagRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	id, err := h.create.Handle(c.Context(), application.CreateFlagRequest{
		Key:          req.Key,
		Description:  req.Description,
		FlagType:     domain.FlagType(req.FlagType),
		DefaultValue: req.DefaultValue,
	})
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, fiber.Map{"id": id})
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateFlagRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	err := h.update.Handle(c.Context(), id, application.UpdateFlagRequest{
		Description:  req.Description,
		DefaultValue: req.DefaultValue,
		Active:       req.Active,
		RolloutPct:   req.RolloutPct,
	})
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, fiber.Map{"updated": true})
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.delete.Handle(c.Context(), id); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"deleted": true})
}

func (h *Handler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	result, err := h.get.Handle(c.Context(), id)
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, result)
}

func (h *Handler) List(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	var active *bool
	if a := c.Query("active"); a != "" {
		v := a == "true"
		active = &v
	}

	result, err := h.list.Handle(c.Context(), domain.FeatureFlagFilter{
		Key:    c.Query("key"),
		Active: active,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, result)
}

func (h *Handler) Evaluate(c *fiber.Ctx) error {
	var req EvaluateFlagRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}

	result, err := h.evaluate.Handle(c.Context(), application.EvaluateRequest{
		Key:        req.Key,
		UserID:     req.UserID,
		Attributes: req.Attributes,
	})
	if err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, result)
}
