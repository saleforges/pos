package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port/repository"
)

var _ PermissionUsecase = (*permUsecase)(nil)

type permUsecase struct {
	permissionRepo repository.PermissionRepository
}

func NewPermissionUsecase(permissionRepo repository.PermissionRepository) *permUsecase {
	return &permUsecase{permissionRepo: permissionRepo}
}

func (uc *permUsecase) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	return uc.permissionRepo.GetAll(ctx)
}

func (uc *permUsecase) CreatePermission(ctx context.Context, permission domain.Permission) error {
	return uc.permissionRepo.Create(ctx, permission)
}

func (uc *permUsecase) DeletePermission(ctx context.Context, permission domain.Permission) error {
	return uc.permissionRepo.Delete(ctx, permission)
}
