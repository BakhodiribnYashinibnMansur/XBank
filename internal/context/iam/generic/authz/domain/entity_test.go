package domain

import (
	"testing"
)

func TestNewRole_Success(t *testing.T) {
	role, err := NewRole("admin", "Administrator role")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role.Name != "admin" {
		t.Errorf("Name expected admin, got: %s", role.Name)
	}
	if role.Description != "Administrator role" {
		t.Errorf("Description mismatch, got: %s", role.Description)
	}
	if role.IsSystem {
		t.Error("new role should not be a system role")
	}
	if role.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestNewRole_EmptyName(t *testing.T) {
	_, err := NewRole("", "No name")
	if err != ErrInvalidRoleName {
		t.Errorf("expected ErrInvalidRoleName, got: %v", err)
	}
}

func TestNewPermission_Success(t *testing.T) {
	perm, err := NewPermission("accounts", "read", "Read accounts")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if perm.Resource != "accounts" {
		t.Errorf("Resource expected accounts, got: %s", perm.Resource)
	}
	if perm.Action != "read" {
		t.Errorf("Action expected read, got: %s", perm.Action)
	}
	if perm.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set")
	}
}

func TestNewPermission_EmptyResource(t *testing.T) {
	_, err := NewPermission("", "read", "desc")
	if err != ErrInvalidResource {
		t.Errorf("expected ErrInvalidResource, got: %v", err)
	}
}

func TestNewPermission_EmptyAction(t *testing.T) {
	_, err := NewPermission("accounts", "", "desc")
	if err != ErrInvalidAction {
		t.Errorf("expected ErrInvalidAction, got: %v", err)
	}
}

func TestNewPolicy_Success(t *testing.T) {
	policy, err := NewPolicy("role-1", "perm-1", ScopeAll)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.RoleID != "role-1" {
		t.Errorf("RoleID expected role-1, got: %s", policy.RoleID)
	}
	if policy.PermissionID != "perm-1" {
		t.Errorf("PermissionID expected perm-1, got: %s", policy.PermissionID)
	}
	if policy.Scope != ScopeAll {
		t.Errorf("Scope expected all, got: %s", policy.Scope)
	}
}

func TestNewPolicy_DefaultScope(t *testing.T) {
	policy, err := NewPolicy("role-1", "perm-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if policy.Scope != ScopeAll {
		t.Errorf("empty scope should default to ScopeAll, got: %s", policy.Scope)
	}
}

func TestNewPolicy_EmptyRoleID(t *testing.T) {
	_, err := NewPolicy("", "perm-1", ScopeOwn)
	if err == nil {
		t.Error("expected error for empty roleID")
	}
}

func TestNewPolicy_EmptyPermissionID(t *testing.T) {
	_, err := NewPolicy("role-1", "", ScopeOwn)
	if err == nil {
		t.Error("expected error for empty permissionID")
	}
}

func TestScopeConstants(t *testing.T) {
	if ScopeOwn != "own" {
		t.Errorf("ScopeOwn expected own, got: %s", ScopeOwn)
	}
	if ScopeAll != "all" {
		t.Errorf("ScopeAll expected all, got: %s", ScopeAll)
	}
}
