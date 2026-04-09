package postgres

import (
	"context"
	"fmt"
	"time"

	domain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/authz/domain"
	sharedpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/db/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/infrastructure/metrics"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WriteRepo struct {
	pool *pgxpool.Pool
}

func NewWriteRepo(pool *pgxpool.Pool) *WriteRepo {
	return &WriteRepo{pool: pool}
}

// ── Roles ───────────────────────────────────────────────────────────

func (r *WriteRepo) CreateRole(ctx context.Context, role *domain.Role) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	err := db.QueryRow(ctx,
		`INSERT INTO rbac_roles (name, description, is_system, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		role.Name, role.Description, role.IsSystem, role.CreatedAt, role.UpdatedAt,
	).Scan(&role.ID)
	metrics.ObserveQuery("AuthzRepo.CreateRole", start, err)
	if err != nil {
		return fmt.Errorf("authz_repo: create_role: %w", err)
	}
	return nil
}

func (r *WriteRepo) GetRoleByID(ctx context.Context, id string) (*domain.Role, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	role := &domain.Role{}
	err := db.QueryRow(ctx,
		`SELECT id, name, description, is_system, created_at, updated_at
		 FROM rbac_roles WHERE id = $1`, id,
	).Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt)
	metrics.ObserveQuery("AuthzRepo.GetRoleByID", start, err)
	if err != nil {
		return nil, domain.ErrRoleNotFound
	}
	return role, nil
}

func (r *WriteRepo) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	role := &domain.Role{}
	err := db.QueryRow(ctx,
		`SELECT id, name, description, is_system, created_at, updated_at
		 FROM rbac_roles WHERE name = $1`, name,
	).Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt)
	metrics.ObserveQuery("AuthzRepo.GetRoleByName", start, err)
	if err != nil {
		return nil, domain.ErrRoleNotFound
	}
	return role, nil
}

func (r *WriteRepo) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, name, description, is_system, created_at, updated_at
		 FROM rbac_roles ORDER BY name`)
	metrics.ObserveQuery("AuthzRepo.ListRoles", start, err)
	if err != nil {
		return nil, fmt.Errorf("authz_repo: list_roles: %w", err)
	}
	defer rows.Close()

	var roles []*domain.Role
	for rows.Next() {
		role := &domain.Role{}
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, fmt.Errorf("authz_repo: list_roles scan: %w", err)
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *WriteRepo) UpdateRole(ctx context.Context, role *domain.Role) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx,
		`UPDATE rbac_roles SET name = $1, description = $2, updated_at = NOW() WHERE id = $3`,
		role.Name, role.Description, role.ID,
	)
	metrics.ObserveQuery("AuthzRepo.UpdateRole", start, err)
	if err != nil {
		return fmt.Errorf("authz_repo: update_role: %w", err)
	}
	return nil
}

func (r *WriteRepo) DeleteRole(ctx context.Context, id string) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM rbac_roles WHERE id = $1`, id)
	metrics.ObserveQuery("AuthzRepo.DeleteRole", start, err)
	if err != nil {
		return fmt.Errorf("authz_repo: delete_role: %w", err)
	}
	return nil
}

// ── Permissions ─────────────────────────────────────────────────────

func (r *WriteRepo) CreatePermission(ctx context.Context, perm *domain.Permission) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	err := db.QueryRow(ctx,
		`INSERT INTO rbac_permissions (resource, action, description, created_at)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		perm.Resource, perm.Action, perm.Description, perm.CreatedAt,
	).Scan(&perm.ID)
	metrics.ObserveQuery("AuthzRepo.CreatePermission", start, err)
	if err != nil {
		return fmt.Errorf("authz_repo: create_permission: %w", err)
	}
	return nil
}

func (r *WriteRepo) GetPermissionByID(ctx context.Context, id string) (*domain.Permission, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	perm := &domain.Permission{}
	err := db.QueryRow(ctx,
		`SELECT id, resource, action, description, created_at
		 FROM rbac_permissions WHERE id = $1`, id,
	).Scan(&perm.ID, &perm.Resource, &perm.Action, &perm.Description, &perm.CreatedAt)
	metrics.ObserveQuery("AuthzRepo.GetPermissionByID", start, err)
	if err != nil {
		return nil, domain.ErrPermissionNotFound
	}
	return perm, nil
}

func (r *WriteRepo) ListPermissions(ctx context.Context) ([]*domain.Permission, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, resource, action, description, created_at
		 FROM rbac_permissions ORDER BY resource, action`)
	metrics.ObserveQuery("AuthzRepo.ListPermissions", start, err)
	if err != nil {
		return nil, fmt.Errorf("authz_repo: list_permissions: %w", err)
	}
	defer rows.Close()

	var perms []*domain.Permission
	for rows.Next() {
		perm := &domain.Permission{}
		if err := rows.Scan(&perm.ID, &perm.Resource, &perm.Action, &perm.Description, &perm.CreatedAt); err != nil {
			return nil, fmt.Errorf("authz_repo: list_permissions scan: %w", err)
		}
		perms = append(perms, perm)
	}
	return perms, nil
}

func (r *WriteRepo) DeletePermission(ctx context.Context, id string) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM rbac_permissions WHERE id = $1`, id)
	metrics.ObserveQuery("AuthzRepo.DeletePermission", start, err)
	if err != nil {
		return fmt.Errorf("authz_repo: delete_permission: %w", err)
	}
	return nil
}

// ── Policies ────────────────────────────────────────────────────────

func (r *WriteRepo) CreatePolicy(ctx context.Context, policy *domain.Policy) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	err := db.QueryRow(ctx,
		`INSERT INTO rbac_policies (role_id, permission_id, scope, created_at)
		 VALUES ($1, $2, $3, $4) RETURNING id`,
		policy.RoleID, policy.PermissionID, policy.Scope, policy.CreatedAt,
	).Scan(&policy.ID)
	metrics.ObserveQuery("AuthzRepo.CreatePolicy", start, err)
	if err != nil {
		return fmt.Errorf("authz_repo: create_policy: %w", err)
	}
	return nil
}

func (r *WriteRepo) DeletePolicy(ctx context.Context, id string) error {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	_, err := db.Exec(ctx, `DELETE FROM rbac_policies WHERE id = $1`, id)
	metrics.ObserveQuery("AuthzRepo.DeletePolicy", start, err)
	if err != nil {
		return fmt.Errorf("authz_repo: delete_policy: %w", err)
	}
	return nil
}

func (r *WriteRepo) ListPoliciesByRole(ctx context.Context, roleID string) ([]*domain.Policy, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)
	rows, err := db.Query(ctx,
		`SELECT id, role_id, permission_id, scope, created_at
		 FROM rbac_policies WHERE role_id = $1 ORDER BY created_at`, roleID)
	metrics.ObserveQuery("AuthzRepo.ListPoliciesByRole", start, err)
	if err != nil {
		return nil, fmt.Errorf("authz_repo: list_policies: %w", err)
	}
	defer rows.Close()

	var policies []*domain.Policy
	for rows.Next() {
		p := &domain.Policy{}
		if err := rows.Scan(&p.ID, &p.RoleID, &p.PermissionID, &p.Scope, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("authz_repo: list_policies scan: %w", err)
		}
		policies = append(policies, p)
	}
	return policies, nil
}

// ── Access Check (hot path) ─────────────────────────────────────────

func (r *WriteRepo) CheckAccess(ctx context.Context, roleName, resource, action string) (*domain.AccessResult, error) {
	start := time.Now()
	db := sharedpg.ExtractDBTX(ctx, r.pool)

	var scope string
	err := db.QueryRow(ctx,
		`SELECT p.scope
		 FROM rbac_policies p
		 JOIN rbac_roles r    ON r.id = p.role_id
		 JOIN rbac_permissions pm ON pm.id = p.permission_id
		 WHERE r.name = $1 AND pm.resource = $2 AND pm.action = $3
		 LIMIT 1`,
		roleName, resource, action,
	).Scan(&scope)
	metrics.ObserveQuery("AuthzRepo.CheckAccess", start, err)

	if err == pgx.ErrNoRows {
		return &domain.AccessResult{Allowed: false}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("authz_repo: check_access: %w", err)
	}

	return &domain.AccessResult{
		Allowed: true,
		Scope:   domain.Scope(scope),
	}, nil
}
