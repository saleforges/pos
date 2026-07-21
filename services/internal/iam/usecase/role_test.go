package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

func TestRoleUsecase_CreateRole(t *testing.T) {
	t.Parallel()

	roleRepo := &mockRoleRepo{}
	uc := NewRoleUsecase(roleRepo, &mockUserRepo{})

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
	_, err = uc.CreateRole(context.Background(), CreateRoleParams{Name: "admin"})
	if !errors.Is(err, domain.ErrInvalidRole) {
		t.Errorf("expected ErrInvalidRole for default role, got %v", err)
	}
}

func TestRoleUsecase_GetRole(t *testing.T) {
	t.Parallel()

	uc := NewRoleUsecase(&mockRoleRepo{
		roles: map[int64]*domain.Role{
			1: {ID: 1, Name: "viewer"},
		},
	}, &mockUserRepo{})

	role, err := uc.GetRole(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role.Name != "viewer" {
		t.Errorf("expected name viewer, got %s", role.Name)
	}

	_, err = uc.GetRole(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for non-existent role")
	}
}

func TestRoleUsecase_ListRoles(t *testing.T) {
	t.Parallel()

	uc := NewRoleUsecase(&mockRoleRepo{}, &mockUserRepo{})
	roles, err := uc.ListRoles(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = roles
}

func TestRoleUsecase_UpdateRole(t *testing.T) {
	t.Parallel()

	desc := "updated description"
	uc := NewRoleUsecase(&mockRoleRepo{
		roles: map[int64]*domain.Role{
			1: {ID: 1, Name: "custom_role", Description: "old description"},
		},
	}, &mockUserRepo{})

	role, err := uc.UpdateRole(context.Background(), UpdateRoleParams{ID: 1, Description: &desc})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role.Description != "updated description" {
		t.Errorf("expected updated description, got %s", role.Description)
	}
}

func TestRoleUsecase_DeleteRole(t *testing.T) {
	t.Parallel()

	uc := NewRoleUsecase(&mockRoleRepo{
		roles: map[int64]*domain.Role{
			1: {ID: 1, Name: "custom_role"},
			2: {ID: 2, Name: "admin", IsSystem: true},
		},
	}, &mockUserRepo{})

	if err := uc.DeleteRole(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := uc.DeleteRole(context.Background(), 999); err == nil {
		t.Fatal("expected error for non-existent role")
	}
}

func TestRoleUsecase_AssignRemoveRole(t *testing.T) {
	t.Parallel()

	uc := NewRoleUsecase(&mockRoleRepo{}, &mockUserRepo{})

	if err := uc.AssignRole(context.Background(), 1, "viewer"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := uc.RemoveRole(context.Background(), 1, "viewer"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRoleUsecase_AssignRemovePermission(t *testing.T) {
	t.Parallel()

	uc := NewRoleUsecase(&mockRoleRepo{}, &mockUserRepo{})

	if err := uc.AssignPermission(context.Background(), 1, domain.UserRead); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := uc.RemovePermission(context.Background(), 1, domain.UserRead); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPermissionUsecase_ListPermissions(t *testing.T) {
	t.Parallel()

	uc := NewPermissionUsecase(&mockPermissionRepo{})
	perms, err := uc.ListPermissions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = perms
}

func TestPermissionUsecase_CreateDeletePermission(t *testing.T) {
	t.Parallel()

	uc := NewPermissionUsecase(&mockPermissionRepo{})

	if err := uc.CreatePermission(context.Background(), domain.Permission("custom.action")); err != nil {
		t.Fatalf("CreatePermission failed: %v", err)
	}
	if err := uc.DeletePermission(context.Background(), domain.Permission("custom.action")); err != nil {
		t.Fatalf("DeletePermission failed: %v", err)
	}
}
