package domain

import (
	"context"
	"time"

	"github.com/BakhodiribnYashinibnMansur/XBank/internal/kernel/domain"
)

// Domain errors
var (
	ErrRoleNotFound       = domain.NewDomainError("ROLE_NOT_FOUND", "role not found")
	ErrPermissionNotFound = domain.NewDomainError("PERMISSION_NOT_FOUND", "permission not found")
	ErrPolicyNotFound     = domain.NewDomainError("POLICY_NOT_FOUND", "policy not found")
	ErrPolicyExists       = domain.NewDomainError("POLICY_EXISTS", "policy already exists for this role and permission")
	ErrSystemRole         = domain.NewDomainError("SYSTEM_ROLE", "system roles cannot be modified or deleted")
	ErrAccessDenied       = domain.NewDomainError("ACCESS_DENIED", "access denied")
	ErrInvalidRoleName    = domain.NewDomainError("INVALID_ROLE_NAME", "role name cannot be empty")
	ErrInvalidResource    = domain.NewDomainError("INVALID_RESOURCE", "resource cannot be empty")
	ErrInvalidAction      = domain.NewDomainError("INVALID_ACTION", "action cannot be empty")
)

// Scope defines the access scope of a policy.
type Scope string

const (
	ScopeOwn Scope = "own"
	ScopeAll Scope = "all"
)

// Role represents an RBAC role.
type Role struct {
	ID          string
	Name        string
	Description string
	IsSystem    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewRole creates a new Role with validation.
func NewRole(name, description string) (*Role, error) {
	if name == "" {
		return nil, ErrInvalidRoleName
	}
	now := time.Now()
	return &Role{
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Permission represents a resource:action permission.
type Permission struct {
	ID          string
	Resource    string
	Action      string
	Description string
	CreatedAt   time.Time
}

// NewPermission creates a new Permission with validation.
func NewPermission(resource, action, description string) (*Permission, error) {
	if resource == "" {
		return nil, ErrInvalidResource
	}
	if action == "" {
		return nil, ErrInvalidAction
	}
	return &Permission{
		Resource:    resource,
		Action:      action,
		Description: description,
		CreatedAt:   time.Now(),
	}, nil
}

// Policy links a Role to a Permission with a Scope.
type Policy struct {
	ID           string
	RoleID       string
	PermissionID string
	Scope        Scope
	CreatedAt    time.Time
}

// NewPolicy creates a new Policy.
func NewPolicy(roleID, permissionID string, scope Scope) (*Policy, error) {
	if roleID == "" || permissionID == "" {
		return nil, ErrPolicyNotFound
	}
	if scope == "" {
		scope = ScopeAll
	}
	return &Policy{
		RoleID:       roleID,
		PermissionID: permissionID,
		Scope:        scope,
		CreatedAt:    time.Now(),
	}, nil
}

// AccessResult is the result of an access check.
type AccessResult struct {
	Allowed bool
	Scope   Scope
}

// Repository defines the persistence contract for RBAC entities.
type Repository interface {
	// Roles
	CreateRole(ctx context.Context, role *Role) error
	GetRoleByID(ctx context.Context, id string) (*Role, error)
	GetRoleByName(ctx context.Context, name string) (*Role, error)
	ListRoles(ctx context.Context) ([]*Role, error)
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, id string) error

	// Permissions
	CreatePermission(ctx context.Context, perm *Permission) error
	GetPermissionByID(ctx context.Context, id string) (*Permission, error)
	ListPermissions(ctx context.Context) ([]*Permission, error)
	DeletePermission(ctx context.Context, id string) error

	// Policies
	CreatePolicy(ctx context.Context, policy *Policy) error
	DeletePolicy(ctx context.Context, id string) error
	ListPoliciesByRole(ctx context.Context, roleID string) ([]*Policy, error)

	// Access check (hot path — single optimized query)
	CheckAccess(ctx context.Context, roleName, resource, action string) (*AccessResult, error)
}
