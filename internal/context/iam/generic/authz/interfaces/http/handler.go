package http

import (
	"net/http"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/authz/application/command"
	authzDomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/authz/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *command.Service
}

func NewHandler(service *command.Service) *Handler {
	return &Handler{service: service}
}

// ── Roles ───────────────────────────────────────────────────────────

func (h *Handler) CreateRole(c *fiber.Ctx) error {
	var req CreateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.Name == "" {
		return apperror.ErrMissingField.WithMessage("name is required")
	}

	role, err := h.service.CreateRole(c.Context(), req.Name, req.Description)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, RoleResponse{
		ID: role.ID, Name: role.Name, Description: role.Description,
		IsSystem: role.IsSystem, CreatedAt: role.CreatedAt, UpdatedAt: role.UpdatedAt,
	})
}

func (h *Handler) UpdateRole(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return apperror.ErrMissingField.WithMessage("id is required")
	}

	var req UpdateRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.Name == "" {
		return apperror.ErrMissingField.WithMessage("name is required")
	}

	role, err := h.service.UpdateRole(c.Context(), id, req.Name, req.Description)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusOK, RoleResponse{
		ID: role.ID, Name: role.Name, Description: role.Description,
		IsSystem: role.IsSystem, CreatedAt: role.CreatedAt, UpdatedAt: role.UpdatedAt,
	})
}

func (h *Handler) DeleteRole(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return apperror.ErrMissingField.WithMessage("id is required")
	}

	if err := h.service.DeleteRole(c.Context(), id); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "Role deleted"})
}

func (h *Handler) ListRoles(c *fiber.Ctx) error {
	roles, err := h.service.ListRoles(c.Context())
	if err != nil {
		return err
	}

	resp := make([]RoleResponse, len(roles))
	for i, r := range roles {
		resp[i] = RoleResponse{
			ID: r.ID, Name: r.Name, Description: r.Description,
			IsSystem: r.IsSystem, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		}
	}
	return apperror.Success(c, http.StatusOK, resp)
}

// ── Permissions ─────────────────────────────────────────────────────

func (h *Handler) CreatePermission(c *fiber.Ctx) error {
	var req CreatePermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.Resource == "" || req.Action == "" {
		return apperror.ErrMissingField.WithMessage("resource and action are required")
	}

	perm, err := h.service.CreatePermission(c.Context(), req.Resource, req.Action, req.Description)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, PermissionResponse{
		ID: perm.ID, Resource: perm.Resource, Action: perm.Action,
		Description: perm.Description, CreatedAt: perm.CreatedAt,
	})
}

func (h *Handler) ListPermissions(c *fiber.Ctx) error {
	perms, err := h.service.ListPermissions(c.Context())
	if err != nil {
		return err
	}

	resp := make([]PermissionResponse, len(perms))
	for i, p := range perms {
		resp[i] = PermissionResponse{
			ID: p.ID, Resource: p.Resource, Action: p.Action,
			Description: p.Description, CreatedAt: p.CreatedAt,
		}
	}
	return apperror.Success(c, http.StatusOK, resp)
}

func (h *Handler) DeletePermission(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return apperror.ErrMissingField.WithMessage("id is required")
	}

	if err := h.service.DeletePermission(c.Context(), id); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "Permission deleted"})
}

// ── Policies ────────────────────────────────────────────────────────

func (h *Handler) AssignPermission(c *fiber.Ctx) error {
	var req AssignPermissionRequest
	if err := c.BodyParser(&req); err != nil {
		return apperror.ErrInvalidJSON
	}
	if req.RoleID == "" || req.PermissionID == "" {
		return apperror.ErrMissingField.WithMessage("role_id and permission_id are required")
	}

	scope := authzDomain.Scope(req.Scope)
	if scope == "" {
		scope = authzDomain.ScopeAll
	}

	policy, err := h.service.AssignPermission(c.Context(), req.RoleID, req.PermissionID, scope)
	if err != nil {
		return err
	}

	return apperror.Success(c, http.StatusCreated, PolicyResponse{
		ID: policy.ID, RoleID: policy.RoleID, PermissionID: policy.PermissionID,
		Scope: string(policy.Scope), CreatedAt: policy.CreatedAt,
	})
}

func (h *Handler) RevokePermission(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return apperror.ErrMissingField.WithMessage("id is required")
	}

	if err := h.service.RevokePermission(c.Context(), id); err != nil {
		return err
	}
	return apperror.Success(c, http.StatusOK, fiber.Map{"message": "Policy revoked"})
}

func (h *Handler) ListPoliciesByRole(c *fiber.Ctx) error {
	roleID := c.Params("role_id")
	if roleID == "" {
		return apperror.ErrMissingField.WithMessage("role_id is required")
	}

	policies, err := h.service.ListPoliciesByRole(c.Context(), roleID)
	if err != nil {
		return err
	}

	resp := make([]PolicyResponse, len(policies))
	for i, p := range policies {
		resp[i] = PolicyResponse{
			ID: p.ID, RoleID: p.RoleID, PermissionID: p.PermissionID,
			Scope: string(p.Scope), CreatedAt: p.CreatedAt,
		}
	}
	return apperror.Success(c, http.StatusOK, resp)
}
