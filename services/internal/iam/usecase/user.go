package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
	"github.com/saleforge/pos/services/internal/iam/port/repository"
	"github.com/saleforge/pos/services/pkg/logger"
	"github.com/saleforge/pos/services/pkg/pagination"
)

type UpdateUserParams struct {
	ID       int64
	Username *string
	Email    *string
	Status   *domain.UserStatus
}

var _ UserUsecase = (*userUsecase)(nil)

type userUsecase struct {
	userRepo       repository.UserRepository
	staffRepo      repository.StaffRepository
	eventPublisher port.EventPublisher
	userCache      port.UserCache
}

func NewUserUsecase(
	userRepo repository.UserRepository,
	staffRepo repository.StaffRepository,
	eventPublisher port.EventPublisher,
	userCache port.UserCache,
) *userUsecase {
	return &userUsecase{
		userRepo:       userRepo,
		staffRepo:      staffRepo,
		eventPublisher: eventPublisher,
		userCache:      userCache,
	}
}

func (uc *userUsecase) ListUsers(ctx context.Context, p pagination.Params) ([]domain.User, *pagination.Metadata, error) {
	data, total, err := uc.userRepo.List(ctx, p.Offset, p.Limit)
	if err != nil {
		return nil, nil, err
	}
	meta := &pagination.Metadata{
		Total:       total,
		Offset:      p.Offset,
		Limit:       p.Limit,
		Count: len(data),
	}
	return data, meta, nil
}

func (uc *userUsecase) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	return uc.userRepo.GetByID(ctx, id)
}

func (uc *userUsecase) UpdateUser(ctx context.Context, input UpdateUserParams) (*domain.User, error) {
	user, err := uc.userRepo.GetByID(ctx, input.ID)
	if err != nil {
		if !errors.Is(err, domain.ErrUserNotFound) {
			logger.Error("update user: get by id failed", "error", err.Error())
			return nil, domain.ErrInternal
		}
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
		logger.Error("update user: update failed", "error", err.Error())
		return nil, domain.ErrInternal
	}
	cacheDel(ctx, uc.userCache, user.ID)
	eventPublish(uc.eventPublisher, ctx, "UserUpdated", map[string]interface{}{"user_id": user.ID})
	return user, nil
}

func (uc *userUsecase) DeleteUser(ctx context.Context, id int64) error {
	if err := uc.userRepo.Delete(ctx, id); err != nil {
		if !errors.Is(err, domain.ErrUserNotFound) {
			logger.Error("delete user: delete failed", "error", err.Error())
			return domain.ErrInternal
		}
		return domain.ErrUserNotFound
	}
	cacheDel(ctx, uc.userCache, id)
	eventPublish(uc.eventPublisher, ctx, "UserDeleted", map[string]interface{}{"user_id": id})
	return nil
}

func (uc *userUsecase) ListStaff(ctx context.Context, userID int64) ([]domain.UserRoleAssignment, error) {
	return uc.staffRepo.ListByUserID(ctx, userID)
}
