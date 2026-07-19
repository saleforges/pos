package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

// RoleUsecase defines role and permission assignment operations.
type RoleUsecase interface {
	ListRoles(ctx context.Context, merchantID *int64) ([]domain.Role, error)
	CreateRole(ctx context.Context, params CreateRoleParams) (*domain.Role, error)
	GetRole(ctx context.Context, id int64) (*domain.Role, error)
	UpdateRole(ctx context.Context, params UpdateRoleParams) (*domain.Role, error)
	DeleteRole(ctx context.Context, id int64) error
	AssignRole(ctx context.Context, userID int64, roleName string) error
	RemoveRole(ctx context.Context, userID int64, roleName string) error
	AssignPermission(ctx context.Context, roleID int64, permission domain.Permission) error
	RemovePermission(ctx context.Context, roleID int64, permission domain.Permission) error
}

// PermissionUsecase defines permission CRUD operations.
type PermissionUsecase interface {
	ListPermissions(ctx context.Context) ([]domain.Permission, error)
	CreatePermission(ctx context.Context, permission domain.Permission) error
	DeletePermission(ctx context.Context, permission domain.Permission) error
}
