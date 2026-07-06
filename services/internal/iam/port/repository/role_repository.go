package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

type RoleRepository interface {
	Create(ctx context.Context, role *domain.Role) error
	GetByName(ctx context.Context, name string) (*domain.Role, error)
	List(ctx context.Context) ([]domain.Role, error)
	Update(ctx context.Context, role *domain.Role) error
	Delete(ctx context.Context, name string) error
	AddPermission(ctx context.Context, roleName string, permission domain.Permission) error
	RemovePermission(ctx context.Context, roleName string, permission domain.Permission) error
	GetPermissions(ctx context.Context, roleName string) ([]domain.Permission, error)
}
