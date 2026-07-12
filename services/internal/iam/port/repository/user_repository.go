package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context, offset, limit int) ([]domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id int64) error
	AddRole(ctx context.Context, userID int64, roleName string) error
	RemoveRole(ctx context.Context, userID int64, roleName string) error
}
