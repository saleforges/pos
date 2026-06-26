package repository

import (
	"context"

	"github.com/saleforge/pos/services/iam/internal/domain"
)

type PermissionRepository interface {
	Create(ctx context.Context, permission domain.Permission) error
	GetAll(ctx context.Context) ([]domain.Permission, error)
	Delete(ctx context.Context, permission domain.Permission) error
}
