package http

import "github.com/gofiber/fiber/v2"

func (h *Handler) RegisterRoutes(group fiber.Router) {
	authz := group.Group("/authz")

	// Roles
	authz.Get("/roles", h.ListRoles)
	authz.Post("/roles", h.CreateRole)
	authz.Put("/roles/:id", h.UpdateRole)
	authz.Delete("/roles/:id", h.DeleteRole)

	// Permissions
	authz.Get("/permissions", h.ListPermissions)
	authz.Post("/permissions", h.CreatePermission)
	authz.Delete("/permissions/:id", h.DeletePermission)

	// Policies
	authz.Get("/roles/:role_id/policies", h.ListPoliciesByRole)
	authz.Post("/policies", h.AssignPermission)
	authz.Delete("/policies/:id", h.RevokePermission)
}
