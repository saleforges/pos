package usecase

import (
	"context"
	"errors"

	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port/repository"
	"github.com/saleforge/pos/services/pkg/logger"
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

type roleUsecase struct {
	roleRepo repository.RoleRepository
	userRepo repository.UserRepository
}

func NewRoleUsecase(roleRepo repository.RoleRepository, userRepo repository.UserRepository) *roleUsecase {
	return &roleUsecase{roleRepo: roleRepo, userRepo: userRepo}
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
		logger.Error("create role: create failed", "error", err.Error())
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
		if !errors.Is(err, domain.ErrInvalidRole) {
			logger.Error("update role: get by id failed", "error", err.Error())
			return nil, domain.ErrInternal
		}
		return nil, domain.ErrInvalidRole
	}
	if input.Description != nil {
		role.Description = *input.Description
	}
	if err := uc.roleRepo.Update(ctx, role); err != nil {
		logger.Error("update role: update failed", "error", err.Error())
		return nil, domain.ErrInternal
	}
	return role, nil
}

func (uc *roleUsecase) DeleteRole(ctx context.Context, id int64) error {
	role, err := uc.roleRepo.GetByID(ctx, id)
	if err != nil {
		if !errors.Is(err, domain.ErrInvalidRole) {
			logger.Error("delete role: get by id failed", "error", err.Error())
			return domain.ErrInternal
		}
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
			if errors.Is(err, domain.ErrInvalidRole) {
				return domain.ErrInvalidRole
			}
			logger.Error("assign role: get by name failed", "error", err.Error())
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
