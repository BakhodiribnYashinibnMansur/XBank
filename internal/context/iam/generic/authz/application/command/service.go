package command

import (
	"context"
	"time"

	authz "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/authz/domain"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
)

type Service struct {
	repo authz.Repository
}

func NewService(repo authz.Repository) *Service {
	return &Service{repo: repo}
}

// ── Roles ───────────────────────────────────────────────────────────

func (s *Service) CreateRole(ctx context.Context, name, description string) (_ *authz.Role, err error) {
	defer metrics.ObserveService("AuthzService", "CreateRole", time.Now(), &err)

	role, err := authz.NewRole(name, description)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateRole(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *Service) UpdateRole(ctx context.Context, id, name, description string) (_ *authz.Role, err error) {
	defer metrics.ObserveService("AuthzService", "UpdateRole", time.Now(), &err)

	role, err := s.repo.GetRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role.IsSystem {
		return nil, authz.ErrSystemRole
	}

	role.Name = name
	role.Description = description
	role.UpdatedAt = time.Now()

	if err := s.repo.UpdateRole(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *Service) DeleteRole(ctx context.Context, id string) (err error) {
	defer metrics.ObserveService("AuthzService", "DeleteRole", time.Now(), &err)

	role, err := s.repo.GetRoleByID(ctx, id)
	if err != nil {
		return err
	}
	if role.IsSystem {
		return authz.ErrSystemRole
	}
	return s.repo.DeleteRole(ctx, id)
}

func (s *Service) ListRoles(ctx context.Context) (_ []*authz.Role, err error) {
	defer metrics.ObserveService("AuthzService", "ListRoles", time.Now(), &err)
	return s.repo.ListRoles(ctx)
}

// ── Permissions ─────────────────────────────────────────────────────

func (s *Service) CreatePermission(ctx context.Context, resource, action, description string) (_ *authz.Permission, err error) {
	defer metrics.ObserveService("AuthzService", "CreatePermission", time.Now(), &err)

	perm, err := authz.NewPermission(resource, action, description)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreatePermission(ctx, perm); err != nil {
		return nil, err
	}
	return perm, nil
}

func (s *Service) ListPermissions(ctx context.Context) (_ []*authz.Permission, err error) {
	defer metrics.ObserveService("AuthzService", "ListPermissions", time.Now(), &err)
	return s.repo.ListPermissions(ctx)
}

func (s *Service) DeletePermission(ctx context.Context, id string) (err error) {
	defer metrics.ObserveService("AuthzService", "DeletePermission", time.Now(), &err)
	return s.repo.DeletePermission(ctx, id)
}

// ── Policies ────────────────────────────────────────────────────────

func (s *Service) AssignPermission(ctx context.Context, roleID, permissionID string, scope authz.Scope) (_ *authz.Policy, err error) {
	defer metrics.ObserveService("AuthzService", "AssignPermission", time.Now(), &err)

	policy, err := authz.NewPolicy(roleID, permissionID, scope)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreatePolicy(ctx, policy); err != nil {
		return nil, err
	}
	return policy, nil
}

func (s *Service) RevokePermission(ctx context.Context, policyID string) (err error) {
	defer metrics.ObserveService("AuthzService", "RevokePermission", time.Now(), &err)
	return s.repo.DeletePolicy(ctx, policyID)
}

func (s *Service) ListPoliciesByRole(ctx context.Context, roleID string) (_ []*authz.Policy, err error) {
	defer metrics.ObserveService("AuthzService", "ListPoliciesByRole", time.Now(), &err)
	return s.repo.ListPoliciesByRole(ctx, roleID)
}

// ── Access Check ────────────────────────────────────────────────────

func (s *Service) CheckAccess(ctx context.Context, roleName, resource, action string) (_ *authz.AccessResult, err error) {
	defer metrics.ObserveService("AuthzService", "CheckAccess", time.Now(), &err)
	return s.repo.CheckAccess(ctx, roleName, resource, action)
}
