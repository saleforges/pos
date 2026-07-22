package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/pkg/pagination"
)

// UserUsecase defines user management operations.
type UserUsecase interface {
	ListUsers(ctx context.Context, p pagination.Params) ([]domain.User, *pagination.Metadata, error)
	GetUser(ctx context.Context, id int64) (*domain.User, error)
	UpdateUser(ctx context.Context, params UpdateUserParams) (*domain.User, error)
	DeleteUser(ctx context.Context, id int64) error
	ListStaff(ctx context.Context, userID int64) ([]domain.UserRoleAssignment, error)
}
