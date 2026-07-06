package memory

import (
	"context"
	"testing"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

func TestUserRepository_CreateAndGetByID(t *testing.T) {
	t.Parallel()
	repo := NewUserRepository()
	user := &domain.User{ID: "u1", Username: "alice", Email: "alice@test.com"}
	user.Roles = []string{"viewer"}

	err := repo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.GetByID(context.Background(), "u1")
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
	_, err := repo.GetByID(context.Background(), "nonexistent")
	if err != domain.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_GetByUsername(t *testing.T) {
	t.Parallel()
	repo := NewUserRepository()
	repo.Create(context.Background(), &domain.User{ID: "u1", Username: "bob", Email: "bob@test.com"})

	got, err := repo.GetByUsername(context.Background(), "bob")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "u1" {
		t.Errorf("expected u1, got %s", got.ID)
	}

	_, err = repo.GetByUsername(context.Background(), "missing")
	if err != domain.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	t.Parallel()
	repo := NewUserRepository()
	repo.Create(context.Background(), &domain.User{ID: "u1", Username: "carol", Email: "carol@test.com"})

	got, err := repo.GetByEmail(context.Background(), "carol@test.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "u1" {
		t.Errorf("expected u1, got %s", got.ID)
	}

	_, err = repo.GetByEmail(context.Background(), "missing@test.com")
	if err != domain.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserRepository_List(t *testing.T) {
	t.Parallel()
	repo := NewUserRepository()
	repo.Create(context.Background(), &domain.User{ID: "u1", Username: "a"})
	repo.Create(context.Background(), &domain.User{ID: "u2", Username: "b"})
	repo.Create(context.Background(), &domain.User{ID: "u3", Username: "c"})

	all, err := repo.List(context.Background(), 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 users, got %d", len(all))
	}

	page, err := repo.List(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page) != 1 {
		t.Errorf("expected 1 user, got %d", len(page))
	}

	empty, err := repo.List(context.Background(), 10, 10)
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
	repo.Create(context.Background(), &domain.User{ID: "u1", Username: "dave"})
	repo.Create(context.Background(), &domain.User{ID: "u1", Username: "dave"})

	err := repo.Update(context.Background(), &domain.User{ID: "u1", Username: "dave"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := repo.GetByID(context.Background(), "u1")
	if got.Username != "dave" {
		t.Errorf("expected dave, got %s", got.Username)
	}
}

func TestUserRepository_Delete(t *testing.T) {
	t.Parallel()
	repo := NewUserRepository()
	repo.Create(context.Background(), &domain.User{ID: "u1", Username: "eve"})

	err := repo.Delete(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = repo.GetByID(context.Background(), "u1")
	if err != domain.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestRoleRepository_GetPermissions(t *testing.T) {
	t.Parallel()
	repo := NewRoleRepository()

	perms, err := repo.GetPermissions(context.Background(), "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(perms) == 0 {
		t.Error("expected non-empty permissions for admin")
	}

	_, err = repo.GetPermissions(context.Background(), "nonexistent")
	if err != domain.ErrInvalidRole {
		t.Errorf("expected ErrInvalidRole, got %v", err)
	}
}

func TestRoleRepository_AssignRole(t *testing.T) {
	t.Parallel()
	userRepo := NewUserRepository()
	userRepo.Create(context.Background(), &domain.User{ID: "u1", Username: "test"})

	err := userRepo.AddRole(context.Background(), "u1", "admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	user, err := userRepo.GetByID(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, r := range user.Roles {
		if r == "admin" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected admin role to be assigned")
	}
}
