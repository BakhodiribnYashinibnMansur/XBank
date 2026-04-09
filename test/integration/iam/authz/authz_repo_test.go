package authz_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	authzdomain "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/authz/domain"
	authzpg "github.com/BakhodiribnYashinibnMansur/XBank/internal/context/iam/generic/authz/infrastructure/postgres"
	"github.com/BakhodiribnYashinibnMansur/XBank/test/integration/common/setup"
)

var pgc *setup.PostgresContainer

func TestMain(m *testing.M) {
	if os.Getenv("INTEGRATION") == "" {
		fmt.Println("skipping integration tests (set INTEGRATION=1 to run)")
		os.Exit(0)
	}

	var code int
	func() {
		t := &testing.T{}
		pgc = setup.MustPostgres(t)
		defer pgc.Teardown()
		code = m.Run()
	}()
	os.Exit(code)
}

func truncate(t *testing.T) {
	t.Helper()
	setup.TruncateAll(t, pgc.Pool)
}

func TestAuthzRepo_CreateRole(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	repo := authzpg.NewWriteRepo(pgc.Pool)

	role, err := authzdomain.NewRole("admin", "Administrator role")
	if err != nil {
		t.Fatalf("NewRole: %v", err)
	}

	if err := repo.CreateRole(ctx, role); err != nil {
		t.Fatalf("repo.CreateRole: %v", err)
	}
	if role.ID == "" {
		t.Fatal("expected role ID to be set after CreateRole")
	}
}

func TestAuthzRepo_GetRoleByID(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	repo := authzpg.NewWriteRepo(pgc.Pool)

	role, _ := authzdomain.NewRole("teller", "Teller role")
	repo.CreateRole(ctx, role)

	got, err := repo.GetRoleByID(ctx, role.ID)
	if err != nil {
		t.Fatalf("repo.GetRoleByID: %v", err)
	}
	if got.Name != "teller" {
		t.Errorf("Name = %q, want %q", got.Name, "teller")
	}
	if got.Description != "Teller role" {
		t.Errorf("Description = %q, want %q", got.Description, "Teller role")
	}
}

func TestAuthzRepo_GetRoleByID_NotFound(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	repo := authzpg.NewWriteRepo(pgc.Pool)

	_, err := repo.GetRoleByID(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error for nonexistent role")
	}
}

func TestAuthzRepo_GetRoleByName(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	repo := authzpg.NewWriteRepo(pgc.Pool)

	role, _ := authzdomain.NewRole("customer", "Customer role")
	repo.CreateRole(ctx, role)

	got, err := repo.GetRoleByName(ctx, "customer")
	if err != nil {
		t.Fatalf("repo.GetRoleByName: %v", err)
	}
	if got.ID != role.ID {
		t.Errorf("ID = %q, want %q", got.ID, role.ID)
	}
}

func TestAuthzRepo_ListRoles(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	repo := authzpg.NewWriteRepo(pgc.Pool)

	r1, _ := authzdomain.NewRole("admin", "Admin")
	repo.CreateRole(ctx, r1)
	r2, _ := authzdomain.NewRole("customer", "Customer")
	repo.CreateRole(ctx, r2)

	roles, err := repo.ListRoles(ctx)
	if err != nil {
		t.Fatalf("repo.ListRoles: %v", err)
	}
	if len(roles) != 2 {
		t.Errorf("len(roles) = %d, want 2", len(roles))
	}
	// Ordered by name: admin, customer
	if roles[0].Name != "admin" {
		t.Errorf("roles[0].Name = %q, want %q", roles[0].Name, "admin")
	}
	if roles[1].Name != "customer" {
		t.Errorf("roles[1].Name = %q, want %q", roles[1].Name, "customer")
	}
}

func TestAuthzRepo_CreatePermission(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	repo := authzpg.NewWriteRepo(pgc.Pool)

	perm, err := authzdomain.NewPermission("accounts", "read", "Read account data")
	if err != nil {
		t.Fatalf("NewPermission: %v", err)
	}

	if err := repo.CreatePermission(ctx, perm); err != nil {
		t.Fatalf("repo.CreatePermission: %v", err)
	}
	if perm.ID == "" {
		t.Fatal("expected permission ID to be set")
	}
}

func TestAuthzRepo_ListPermissions(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	repo := authzpg.NewWriteRepo(pgc.Pool)

	p1, _ := authzdomain.NewPermission("accounts", "read", "Read accounts")
	repo.CreatePermission(ctx, p1)
	p2, _ := authzdomain.NewPermission("accounts", "write", "Write accounts")
	repo.CreatePermission(ctx, p2)
	p3, _ := authzdomain.NewPermission("transfers", "create", "Create transfers")
	repo.CreatePermission(ctx, p3)

	perms, err := repo.ListPermissions(ctx)
	if err != nil {
		t.Fatalf("repo.ListPermissions: %v", err)
	}
	if len(perms) != 3 {
		t.Errorf("len(perms) = %d, want 3", len(perms))
	}
}

func TestAuthzRepo_CreatePolicy(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	repo := authzpg.NewWriteRepo(pgc.Pool)

	role, _ := authzdomain.NewRole("admin", "Admin")
	repo.CreateRole(ctx, role)

	perm, _ := authzdomain.NewPermission("accounts", "read", "Read accounts")
	repo.CreatePermission(ctx, perm)

	policy, err := authzdomain.NewPolicy(role.ID, perm.ID, authzdomain.ScopeAll)
	if err != nil {
		t.Fatalf("NewPolicy: %v", err)
	}

	if err := repo.CreatePolicy(ctx, policy); err != nil {
		t.Fatalf("repo.CreatePolicy: %v", err)
	}
	if policy.ID == "" {
		t.Fatal("expected policy ID to be set")
	}
}

func TestAuthzRepo_CheckAccess_Allowed(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	repo := authzpg.NewWriteRepo(pgc.Pool)

	// Setup: role + permission + policy
	role, _ := authzdomain.NewRole("admin", "Admin")
	repo.CreateRole(ctx, role)

	perm, _ := authzdomain.NewPermission("accounts", "read", "Read accounts")
	repo.CreatePermission(ctx, perm)

	policy, _ := authzdomain.NewPolicy(role.ID, perm.ID, authzdomain.ScopeAll)
	repo.CreatePolicy(ctx, policy)

	result, err := repo.CheckAccess(ctx, "admin", "accounts", "read")
	if err != nil {
		t.Fatalf("repo.CheckAccess: %v", err)
	}
	if !result.Allowed {
		t.Error("expected access to be allowed")
	}
	if result.Scope != authzdomain.ScopeAll {
		t.Errorf("Scope = %q, want %q", result.Scope, authzdomain.ScopeAll)
	}
}

func TestAuthzRepo_CheckAccess_Denied(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	repo := authzpg.NewWriteRepo(pgc.Pool)

	result, err := repo.CheckAccess(ctx, "nonexistent", "accounts", "read")
	if err != nil {
		t.Fatalf("repo.CheckAccess: %v", err)
	}
	if result.Allowed {
		t.Error("expected access to be denied")
	}
}

func TestAuthzRepo_CheckAccess_OwnScope(t *testing.T) {
	truncate(t)
	ctx := context.Background()
	repo := authzpg.NewWriteRepo(pgc.Pool)

	role, _ := authzdomain.NewRole("customer", "Customer")
	repo.CreateRole(ctx, role)

	perm, _ := authzdomain.NewPermission("accounts", "read", "Read own accounts")
	repo.CreatePermission(ctx, perm)

	policy, _ := authzdomain.NewPolicy(role.ID, perm.ID, authzdomain.ScopeOwn)
	repo.CreatePolicy(ctx, policy)

	result, err := repo.CheckAccess(ctx, "customer", "accounts", "read")
	if err != nil {
		t.Fatalf("repo.CheckAccess: %v", err)
	}
	if !result.Allowed {
		t.Error("expected access to be allowed with 'own' scope")
	}
	if result.Scope != authzdomain.ScopeOwn {
		t.Errorf("Scope = %q, want %q", result.Scope, authzdomain.ScopeOwn)
	}
}
