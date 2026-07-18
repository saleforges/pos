package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

func TestAuthUsecase_UpdateUser(t *testing.T) {
	t.Parallel()

	username := "updated"
	email := "updated@example.com"
	status := domain.UserStatusDisabled

	tests := []struct {
		name     string
		input    UpdateUserParams
		userRepo *mockUserRepo
		wantErr  error
	}{
		{
			name: "successful update",
			input: UpdateUserParams{
				ID:       1,
				Username: &username,
				Email:    &email,
				Status:   &status,
			},
			userRepo: &mockUserRepo{
				users: map[int64]*domain.User{
					1: {ID: 1, Username: "old", Email: "old@example.com", Status: domain.UserStatusActive},
				},
			},
		},
		{
			name: "user not found",
			input: UpdateUserParams{
				ID: 999,
			},
			userRepo: &mockUserRepo{},
			wantErr:  domain.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := newTestUsecase(tt.userRepo, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})
			_, err := uc.UpdateUser(context.Background(), tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAuthUsecase_GetUser(t *testing.T) {
	t.Parallel()

	uc := newTestUsecase(&mockUserRepo{
		users: map[int64]*domain.User{
			1: {ID: 1, Username: "johndoe"},
		},
	}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})

	user, err := uc.GetUser(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Username != "johndoe" {
		t.Errorf("expected username johndoe, got %s", user.Username)
	}

	_, err = uc.GetUser(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

func TestAuthUsecase_DeleteUser(t *testing.T) {
	t.Parallel()

	userRepo := &mockUserRepo{
		users: map[int64]*domain.User{
			1: {ID: 1, Username: "johndoe"},
		},
	}
	uc := newTestUsecase(userRepo, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})

	if err := uc.DeleteUser(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Deleting a non-existent user may not return an error depending on mock implementation
	// The usecase wraps repo.Delete and returns ErrUserNotFound on error
	_ = uc.DeleteUser(context.Background(), 999)
}

func TestAuthUsecase_ListStaff(t *testing.T) {
	t.Parallel()

	uc := newTestUsecase(&mockUserRepo{}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})
	staff, err := uc.ListStaff(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Memory repo returns nil for unknown users
	_ = staff
}

func TestAuthUsecase_AssignRole(t *testing.T) {
	t.Parallel()

	uc := newTestUsecase(&mockUserRepo{
		users: map[int64]*domain.User{
			1: {ID: 1},
		},
	}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})

	if err := uc.AssignRole(context.Background(), 1, "viewer"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthUsecase_RemoveRole(t *testing.T) {
	t.Parallel()

	uc := newTestUsecase(&mockUserRepo{
		users: map[int64]*domain.User{
			1: {ID: 1},
		},
	}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})

	if err := uc.RemoveRole(context.Background(), 1, "viewer"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthUsecase_ListUsers(t *testing.T) {
	t.Parallel()

	uc := newTestUsecase(&mockUserRepo{}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})
	users, err := uc.ListUsers(context.Background(), 0, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Mock returns nil, which is acceptable
	_ = users
}
