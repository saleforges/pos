package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

// UserUsecase defines user management operations.
type UserUsecase interface {
	ListUsers(ctx context.Context, offset, limit int) ([]domain.User, error)
	GetUser(ctx context.Context, id int64) (*domain.User, error)
	UpdateUser(ctx context.Context, params UpdateUserParams) (*domain.User, error)
	DeleteUser(ctx context.Context, id int64) error
	ListStaff(ctx context.Context, userID int64) ([]domain.UserRoleAssignment, error)
}
