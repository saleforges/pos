package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

func TestAuthUsecase_CreateRole(t *testing.T) {
	t.Parallel()

	uc := newTestUsecase(&mockUserRepo{}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})

	role, err := uc.CreateRole(context.Background(), CreateRoleParams{
		Name:        "custom_role",
		Description: "A custom role",
		Permissions: []domain.Permission{domain.UserRead},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role.Name != "custom_role" {
		t.Errorf("expected name custom_role, got %s", role.Name)
	}

	// Creating a default role should fail
	_, err = uc.CreateRole(context.Background(), CreateRoleParams{
		Name: "admin",
	})
	if !errors.Is(err, domain.ErrInvalidRole) {
		t.Errorf("expected ErrInvalidRole for default role, got %v", err)
	}
}

func TestAuthUsecase_GetRole(t *testing.T) {
	t.Parallel()

	uc := newTestUsecase(&mockUserRepo{}, &mockRoleRepo{
		roles: map[int64]*domain.Role{
			1: {ID: 1, Name: "viewer"},
		},
	}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})

	role, err := uc.GetRole(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role.Name != "viewer" {
		t.Errorf("expected role name viewer, got %s", role.Name)
	}

	_, err = uc.GetRole(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for non-existent role")
	}
}

func TestAuthUsecase_ListRoles(t *testing.T) {
	t.Parallel()

	uc := newTestUsecase(&mockUserRepo{}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})
	roles, err := uc.ListRoles(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Mock returns nil slice, which is acceptable
	if roles != nil {
		t.Logf("got %d roles", len(roles))
	}
}

func TestAuthUsecase_UpdateRole(t *testing.T) {
	t.Parallel()

	desc := "updated description"
	uc := newTestUsecase(&mockUserRepo{}, &mockRoleRepo{
		roles: map[int64]*domain.Role{
			1: {ID: 1, Name: "custom_role", Description: "old description"},
		},
	}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})

	role, err := uc.UpdateRole(context.Background(), UpdateRoleParams{
		ID:          1,
		Description: &desc,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role.Description != "updated description" {
		t.Errorf("expected updated description, got %s", role.Description)
	}
}

func TestAuthUsecase_DeleteRole(t *testing.T) {
	t.Parallel()

	uc := newTestUsecase(&mockUserRepo{}, &mockRoleRepo{
		roles: map[int64]*domain.Role{
			1: {ID: 1, Name: "custom_role"},
			2: {ID: 2, Name: "admin", IsSystem: true},
		},
	}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})

	// Deleting a custom role should succeed
	if err := uc.DeleteRole(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Deleting a non-existent role should fail
	if err := uc.DeleteRole(context.Background(), 999); err == nil {
		t.Fatal("expected error for non-existent role")
	}
}

func TestAuthUsecase_AssignPermission(t *testing.T) {
	t.Parallel()

	uc := newTestUsecase(&mockUserRepo{}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})
	if err := uc.AssignPermission(context.Background(), 1, domain.UserRead); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthUsecase_RemovePermission(t *testing.T) {
	t.Parallel()

	uc := newTestUsecase(&mockUserRepo{}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})
	if err := uc.RemovePermission(context.Background(), 1, domain.UserRead); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthUsecase_ListPermissions(t *testing.T) {
	t.Parallel()

	uc := newTestUsecase(&mockUserRepo{}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})
	perms, err := uc.ListPermissions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Mock returns nil, which is acceptable
	if perms != nil {
		t.Logf("got %d permissions", len(perms))
	}
}

func TestAuthUsecase_CreateDeletePermission(t *testing.T) {
	t.Parallel()

	uc := newTestUsecase(&mockUserRepo{}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})

	err := uc.CreatePermission(context.Background(), domain.Permission("custom.action"))
	if err != nil {
		t.Fatalf("CreatePermission failed: %v", err)
	}

	err = uc.DeletePermission(context.Background(), domain.Permission("custom.action"))
	if err != nil {
		t.Fatalf("DeletePermission failed: %v", err)
	}
}
