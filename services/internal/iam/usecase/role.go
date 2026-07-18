package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

type CreateRoleParams struct {
	Name        string
	Description string
	Permissions []domain.Permission
}

type UpdateRoleParams struct {
	ID          int64
	Description *string
}

func (uc *authUsecase) ListRoles(ctx context.Context, merchantID *int64) ([]domain.Role, error) {
	return uc.roleRepo.List(ctx, merchantID)
}

func (uc *authUsecase) CreateRole(ctx context.Context, input CreateRoleParams) (*domain.Role, error) {
	if _, ok := domain.DefaultRoles[input.Name]; ok {
		return nil, domain.ErrInvalidRole
	}

	role := &domain.Role{
		Name:        input.Name,
		Description: input.Description,
		Permissions: input.Permissions,
	}

	if err := uc.roleRepo.Create(ctx, role); err != nil {
		return nil, domain.ErrInternal
	}

	return role, nil
}

func (uc *authUsecase) GetRole(ctx context.Context, id int64) (*domain.Role, error) {
	return uc.roleRepo.GetByID(ctx, id)
}

func (uc *authUsecase) UpdateRole(ctx context.Context, input UpdateRoleParams) (*domain.Role, error) {
	role, err := uc.roleRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, domain.ErrInvalidRole
	}

	if input.Description != nil {
		role.Description = *input.Description
	}

	if err := uc.roleRepo.Update(ctx, role); err != nil {
		return nil, domain.ErrInternal
	}

	return role, nil
}

func (uc *authUsecase) DeleteRole(ctx context.Context, id int64) error {
	role, err := uc.roleRepo.GetByID(ctx, id)
	if err != nil {
		return domain.ErrInvalidRole
	}
	if _, ok := domain.DefaultRoles[role.Name]; ok {
		return domain.ErrInvalidRole
	}

	return uc.roleRepo.Delete(ctx, role.ID)
}

func (uc *authUsecase) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	return uc.permissionRepo.GetAll(ctx)
}

func (uc *authUsecase) CreatePermission(ctx context.Context, permission domain.Permission) error {
	return uc.permissionRepo.Create(ctx, permission)
}

func (uc *authUsecase) DeletePermission(ctx context.Context, permission domain.Permission) error {
	return uc.permissionRepo.Delete(ctx, permission)
}

func (uc *authUsecase) AssignPermission(ctx context.Context, roleID int64, permission domain.Permission) error {
	return uc.roleRepo.AddPermission(ctx, roleID, permission)
}

func (uc *authUsecase) RemovePermission(ctx context.Context, roleID int64, permission domain.Permission) error {
	return uc.roleRepo.RemovePermission(ctx, roleID, permission)
}
