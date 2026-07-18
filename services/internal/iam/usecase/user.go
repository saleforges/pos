package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

type UpdateUserParams struct {
	ID       int64
	Username *string
	Email    *string
	Status   *domain.UserStatus
}

func (uc *authUsecase) ListUsers(ctx context.Context, offset, limit int) ([]domain.User, error) {
	return uc.userRepo.List(ctx, offset, limit)
}

func (uc *authUsecase) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}

func (uc *authUsecase) UpdateUser(ctx context.Context, input UpdateUserParams) (*domain.User, error) {
	user, err := uc.userRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, domain.ErrUserNotFound
	}

	if input.Username != nil {
		user.Username = *input.Username
	}
	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Status != nil {
		user.Status = *input.Status
	}

	user.UpdatedAt = time.Now().UTC()

	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, domain.ErrInternal
	}

	uc.cacheDel(ctx, user.ID)

	uc.publishEvent(ctx, "UserUpdated", map[string]interface{}{
		"user_id": user.ID,
	})

	return user, nil
}

func (uc *authUsecase) DeleteUser(ctx context.Context, id int64) error {
	if err := uc.userRepo.Delete(ctx, id); err != nil {
		return domain.ErrUserNotFound
	}

	uc.cacheDel(ctx, id)

	uc.publishEvent(ctx, "UserDeleted", map[string]interface{}{
		"user_id": id,
	})

	return nil
}

func (uc *authUsecase) AssignRole(ctx context.Context, userID int64, roleName string) error {
	if _, ok := domain.DefaultRoles[roleName]; !ok {
		if _, err := uc.roleRepo.GetByName(ctx, roleName); err != nil {
			return domain.ErrInvalidRole
		}
	}

	if err := uc.userRepo.AddRole(ctx, userID, roleName); err != nil {
		return err
	}

	uc.cacheDel(ctx, userID)

	uc.publishEvent(ctx, "RoleAssigned", map[string]interface{}{
		"user_id": userID,
		"role":    roleName,
	})

	return nil
}

func (uc *authUsecase) RemoveRole(ctx context.Context, userID int64, roleName string) error {
	if err := uc.userRepo.RemoveRole(ctx, userID, roleName); err != nil {
		return err
	}

	uc.cacheDel(ctx, userID)

	uc.publishEvent(ctx, "RoleRevoked", map[string]interface{}{
		"user_id": userID,
		"role":    roleName,
	})

	return nil
}

func (uc *authUsecase) ListStaff(ctx context.Context, userID int64) ([]domain.UserRoleAssignment, error) {
	return uc.staffRepo.ListByUserID(ctx, userID)
}
