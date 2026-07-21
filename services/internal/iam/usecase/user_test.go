package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

func TestUserUsecase_UpdateUser(t *testing.T) {
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
			uc := NewUserUsecase(tt.userRepo, &mockStaffRepo{}, &mockEventPublisher{}, nil)
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

func TestUserUsecase_GetUser(t *testing.T) {
	t.Parallel()

	uc := NewUserUsecase(&mockUserRepo{
		users: map[int64]*domain.User{
			1: {ID: 1, Username: "johndoe"},
		},
	}, &mockStaffRepo{}, &mockEventPublisher{}, nil)

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

func TestUserUsecase_DeleteUser(t *testing.T) {
	t.Parallel()

	userRepo := &mockUserRepo{
		users: map[int64]*domain.User{
			1: {ID: 1, Username: "johndoe"},
		},
	}
	uc := NewUserUsecase(userRepo, &mockStaffRepo{}, &mockEventPublisher{}, nil)

	if err := uc.DeleteUser(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserUsecase_ListStaff(t *testing.T) {
	t.Parallel()

	uc := NewUserUsecase(&mockUserRepo{}, &mockStaffRepo{}, &mockEventPublisher{}, nil)
	staff, err := uc.ListStaff(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = staff
}

func TestUserUsecase_ListUsers(t *testing.T) {
	t.Parallel()

	uc := NewUserUsecase(&mockUserRepo{}, &mockStaffRepo{}, &mockEventPublisher{}, nil)
	users, err := uc.ListUsers(context.Background(), 0, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = users
}

type mockStaffRepo struct{}

func (m *mockStaffRepo) ListByUserID(_ context.Context, userID int64) ([]domain.UserRoleAssignment, error) {
	return nil, nil
}
func (m *mockStaffRepo) SetDefaultRole(_ context.Context, userID, roleID int64) error { return nil }
func (m *mockStaffRepo) Create(_ context.Context, userID int64, merchantID int64, merchantName, role string) error {
	return nil
}
