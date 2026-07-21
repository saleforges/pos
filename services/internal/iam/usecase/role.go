package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port/repository"
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

var _ RoleUsecase = (*roleUsecase)(nil)
var _ PermissionUsecase = (*permUsecase)(nil)

type roleUsecase struct {
	roleRepo repository.RoleRepository
	userRepo repository.UserRepository
}

type permUsecase struct {
	permissionRepo repository.PermissionRepository
}

func NewRoleUsecase(roleRepo repository.RoleRepository, userRepo repository.UserRepository) *roleUsecase {
	return &roleUsecase{roleRepo: roleRepo, userRepo: userRepo}
}

func NewPermissionUsecase(permissionRepo repository.PermissionRepository) *permUsecase {
	return &permUsecase{permissionRepo: permissionRepo}
}

func (uc *roleUsecase) ListRoles(ctx context.Context, merchantID *int64) ([]domain.Role, error) {
	return uc.roleRepo.List(ctx, merchantID)
}

func (uc *roleUsecase) CreateRole(ctx context.Context, input CreateRoleParams) (*domain.Role, error) {
	if _, ok := domain.DefaultRoles[input.Name]; ok {
		return nil, domain.ErrInvalidRole
	}
	role := &domain.Role{Name: input.Name, Description: input.Description, Permissions: input.Permissions}
	if err := uc.roleRepo.Create(ctx, role); err != nil {
		return nil, domain.ErrInternal
	}
	return role, nil
}

func (uc *roleUsecase) GetRole(ctx context.Context, id int64) (*domain.Role, error) {
	return uc.roleRepo.GetByID(ctx, id)
}

func (uc *roleUsecase) UpdateRole(ctx context.Context, input UpdateRoleParams) (*domain.Role, error) {
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

func (uc *roleUsecase) DeleteRole(ctx context.Context, id int64) error {
	role, err := uc.roleRepo.GetByID(ctx, id)
	if err != nil {
		return domain.ErrInvalidRole
	}
	if _, ok := domain.DefaultRoles[role.Name]; ok {
		return domain.ErrInvalidRole
	}
	return uc.roleRepo.Delete(ctx, role.ID)
}

func (uc *roleUsecase) AssignRole(ctx context.Context, userID int64, roleName string) error {
	if _, ok := domain.DefaultRoles[roleName]; !ok {
		if _, err := uc.roleRepo.GetByName(ctx, roleName); err != nil {
			return domain.ErrInvalidRole
		}
	}
	return uc.userRepo.AddRole(ctx, userID, roleName)
}

func (uc *roleUsecase) RemoveRole(ctx context.Context, userID int64, roleName string) error {
	return uc.userRepo.RemoveRole(ctx, userID, roleName)
}

func (uc *roleUsecase) AssignPermission(ctx context.Context, roleID int64, permission domain.Permission) error {
	return uc.roleRepo.AddPermission(ctx, roleID, permission)
}

func (uc *roleUsecase) RemovePermission(ctx context.Context, roleID int64, permission domain.Permission) error {
	return uc.roleRepo.RemovePermission(ctx, roleID, permission)
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
