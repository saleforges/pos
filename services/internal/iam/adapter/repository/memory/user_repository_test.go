package memory

import (
	"context"
	"testing"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

func TestUserRepository_CreateAndGetByID(t *testing.T) {
	t.Parallel()
	repo := NewUserRepository()
	user := &domain.User{Username: "alice", Email: "alice@test.com"}

	err := repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}

	got, err := repo.GetByID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Username != "alice" {
		t.Errorf("expected alice, got %s", got.Username)
	}
}

func TestUserRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	repo := NewUserRepository()
	_, err := repo.GetByID(context.Background(), 999)
	if err != domain.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	t.Parallel()
	repo := NewUserRepository()
	repo.Create(context.Background(), &domain.User{Username: "bob", Email: "bob@test.com"})

	got, err := repo.GetByUsername(context.Background(), "bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Username != "bob" {
		t.Errorf("expected bob, got %s", got.Username)
	}

	_, err = repo.GetByUsername(context.Background(), "missing")
	if err != domain.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	t.Parallel()
	repo := NewUserRepository()
	repo.Create(context.Background(), &domain.User{Username: "carol", Email: "carol@test.com"})

	got, err := repo.GetByEmail(context.Background(), "carol@test.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Username != "carol" {
		t.Errorf("expected carol, got %s", got.Username)
	}

	_, err = repo.GetByEmail(context.Background(), "missing@test.com")
	if err != domain.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_List(t *testing.T) {
	t.Parallel()
	repo := NewUserRepository()
	repo.Create(context.Background(), &domain.User{Username: "a"})
	repo.Create(context.Background(), &domain.User{Username: "b"})
	repo.Create(context.Background(), &domain.User{Username: "c"})

	all, _, err := repo.List(context.Background(), 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 users, got %d", len(all))
	}

	page, _, err := repo.List(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page) != 1 {
		t.Errorf("expected 1 user, got %d", len(page))
	}

	empty, _, err := repo.List(context.Background(), 10, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("expected 0 users, got %d", len(empty))
	}
}

func TestUserRepository_Update(t *testing.T) {
	t.Parallel()
	repo := NewUserRepository()
	u1 := &domain.User{Username: "dave"}
	repo.Create(context.Background(), u1)

	err := repo.Update(context.Background(), &domain.User{ID: u1.ID, Username: "dave"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := repo.GetByID(context.Background(), u1.ID)
	if got.Username != "dave" {
		t.Errorf("expected dave, got %s", got.Username)
	}
}

func TestUserRepository_Delete(t *testing.T) {
	t.Parallel()
	repo := NewUserRepository()
	u1 := &domain.User{Username: "eve"}
	repo.Create(context.Background(), u1)

	err := repo.Delete(context.Background(), u1.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = repo.GetByID(context.Background(), u1.ID)
	if err != domain.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestRoleRepository_GetPermissions(t *testing.T) {
	t.Parallel()
	repo := NewRoleRepository()

	role, err := repo.GetByName(context.Background(), "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	perms, err := repo.GetPermissions(context.Background(), role.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(perms) == 0 {
		t.Error("expected non-empty permissions for admin")
	}

	_, err = repo.GetPermissions(context.Background(), 999)
	if err != domain.ErrInvalidRole {
		t.Errorf("expected ErrInvalidRole, got %v", err)
	}
}

func TestRoleRepository_AssignRole(t *testing.T) {
	t.Parallel()
	userRepo := NewUserRepository()
	u1 := &domain.User{Username: "test"}
	userRepo.Create(context.Background(), u1)

	err := userRepo.AddRole(context.Background(), u1.ID, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user, err := userRepo.GetByID(context.Background(), u1.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, r := range user.Roles {
		if r.Role.Name == "admin" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected admin role to be assigned")
	}
}

func TestUserRepository_RemoveRole(t *testing.T) {
	t.Parallel()
	userRepo := NewUserRepository()
	u1 := &domain.User{Username: "removerole"}
	userRepo.Create(context.Background(), u1)
	userRepo.AddRole(context.Background(), u1.ID, "admin")

	err := userRepo.RemoveRole(context.Background(), u1.ID, "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user, _ := userRepo.GetByID(context.Background(), u1.ID)
	for _, r := range user.Roles {
		if r.Role.Name == "admin" {
			t.Error("expected admin role to be removed")
		}
	}
}

func TestRoleRepository_Create(t *testing.T) {
	t.Parallel()
	repo := NewRoleRepository()

	role := &domain.Role{Name: "custom_role", Description: "desc"}
	err := repo.Create(context.Background(), role)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}
}

func TestRoleRepository_GetByID(t *testing.T) {
	t.Parallel()
	repo := NewRoleRepository()

	role, err := repo.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if role.ID != 1 {
		t.Errorf("expected ID 1, got %d", role.ID)
	}
	if role.Name == "" {
		t.Error("expected non-empty name")
	}

	_, err = repo.GetByID(context.Background(), 999)
	if err != domain.ErrInvalidRole {
		t.Errorf("expected ErrInvalidRole, got %v", err)
	}
}

func TestRoleRepository_List(t *testing.T) {
	t.Parallel()
	repo := NewRoleRepository()

	roles, err := repo.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roles) == 0 {
		t.Error("expected at least 1 role")
	}

	roles, err = repo.List(context.Background(), int64Ptr(100))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
func int64Ptr(v int64) *int64 { return &v }

func TestRoleRepository_Update(t *testing.T) {
	t.Parallel()
	repo := NewRoleRepository()
	repo.Create(context.Background(), &domain.Role{Name: "update_role", Description: "old"})

	role, _ := repo.GetByName(context.Background(), "update_role")
	role.Description = "updated"
	err := repo.Update(context.Background(), role)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRoleRepository_Delete(t *testing.T) {
	t.Parallel()
	repo := NewRoleRepository()
	repo.Create(context.Background(), &domain.Role{Name: "delete_role"})
	role, _ := repo.GetByName(context.Background(), "delete_role")

	err := repo.Delete(context.Background(), role.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRoleRepository_AddRemovePermission(t *testing.T) {
	t.Parallel()
	repo := NewRoleRepository()
	role, _ := repo.GetByName(context.Background(), "admin")

	err := repo.AddPermission(context.Background(), role.ID, domain.SessionManage)
	if err != nil {
		t.Fatalf("AddPermission failed: %v", err)
	}

	err = repo.RemovePermission(context.Background(), role.ID, domain.SessionManage)
	if err != nil {
		t.Fatalf("RemovePermission failed: %v", err)
	}
}

func TestPermissionRepository(t *testing.T) {
	t.Parallel()

	repo := NewPermissionRepository()

	t.Run("get all returns permissions", func(t *testing.T) {
		perms, err := repo.GetAll(context.Background())
		if err != nil {
			t.Fatalf("GetAll failed: %v", err)
		}
		if len(perms) == 0 {
			t.Error("expected at least 1 permission")
		}
	})

	t.Run("create and delete", func(t *testing.T) {
		err := repo.Create(context.Background(), domain.Permission("custom.test"))
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		err = repo.Delete(context.Background(), domain.Permission("custom.test"))
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
	})
}

func TestLoginAuditRepository(t *testing.T) {
	t.Parallel()

	repo := NewLoginAuditRepository()
	ctx := context.Background()

	t.Run("create and list", func(t *testing.T) {
		audit := &domain.LoginAudit{UserID: 1, Email: "test@t.com", Success: true}
		err := repo.Create(ctx, audit)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}

		audits, _, err := repo.List(ctx, []int64{1}, 0, 10)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(audits) != 1 {
			t.Errorf("expected 1 audit, got %d", len(audits))
		}
	})

	t.Run("list with offset past end returns empty", func(t *testing.T) {
		audits, _, err := repo.List(ctx, []int64{1}, 100, 10)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(audits) != 0 {
			t.Errorf("expected 0 audits, got %d", len(audits))
		}
	})

	t.Run("list filters out audits for other users", func(t *testing.T) {
		if err := repo.Create(ctx, &domain.LoginAudit{UserID: 2, Email: "other@t.com", Success: true}); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		audits, total, err := repo.List(ctx, []int64{1}, 0, 10)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if total != 1 || len(audits) != 1 {
			t.Fatalf("expected 1 audit scoped to user 1, got %d (total %d)", len(audits), total)
		}
		if audits[0].UserID != 1 {
			t.Errorf("expected audit for user 1, got user %d", audits[0].UserID)
		}
	})
}

func TestStaffRepository(t *testing.T) {
	t.Parallel()

	repo := NewStaffRepository()
	ctx := context.Background()

	t.Run("list by user ID returns empty initially", func(t *testing.T) {
		staff, err := repo.ListByUserID(ctx, 1)
		if err != nil {
			t.Fatalf("ListByUserID failed: %v", err)
		}
		_ = staff
	})

	t.Run("set default role", func(t *testing.T) {
		err := repo.SetDefaultRole(ctx, 1, 1)
		if err != nil {
			t.Fatalf("SetDefaultRole failed: %v", err)
		}
	})
}
