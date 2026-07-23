package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/pkg/pagination"
)

type AssignStaffParams struct {
	MerchantID int64
	BranchID   int64
	UserID     int64
	Role       domain.StaffRole
	IsDefault  bool
}

type UpdateStaffParams struct {
	ID     int64
	Role   *domain.StaffRole
	Status *domain.StaffStatus
}

func (uc *merchantUsecase) AssignStaff(ctx context.Context, input AssignStaffParams) (*domain.StaffMember, error) {
	if input.BranchID == 0 || input.UserID == 0 || input.Role == "" {
		return nil, domain.ErrInvalidStaff
	}

	_, err := uc.merchantRepo.GetByID(ctx, input.MerchantID)
	if err != nil {
		return nil, domain.ErrMerchantNotFound
	}

	_, err = uc.branchRepo.GetByID(ctx, input.BranchID)
	if err != nil {
		return nil, domain.ErrBranchNotFound
	}

	existing, err := uc.staffRepo.GetByUserAndBranch(ctx, input.UserID, input.BranchID)
	if err != nil && !errors.Is(err, domain.ErrStaffNotFound) {
		return nil, domain.ErrInternal
	}
	if existing != nil {
		return nil, domain.ErrStaffExists
	}

	now := time.Now().UTC()
	staff := &domain.StaffMember{
		MerchantID: input.MerchantID,
		BranchID:   input.BranchID,
		UserID:     input.UserID,
		Role:       input.Role,
		Status:     domain.StaffStatusActive,
		IsDefault:  input.IsDefault,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if input.IsDefault {
		if err := uc.staffRepo.SetDefaultBranch(ctx, input.UserID, input.BranchID); err != nil {
			return nil, fmt.Errorf("failed to set default branch: %w", err)
		}
	}

	if err := uc.staffRepo.Create(ctx, staff); err != nil {
		if errors.Is(err, domain.ErrStaffExists) {
			return nil, err
		}
		return nil, domain.ErrInternal
	}
	return staff, nil
}

func (uc *merchantUsecase) GetStaff(ctx context.Context, id int64) (*domain.StaffMember, error) {
	return uc.staffRepo.GetByID(ctx, id)
}

func (uc *merchantUsecase) ListStaffByBranch(ctx context.Context, branchID int64, p pagination.Params) ([]domain.StaffMember, *pagination.Metadata, error) {
	data, total, err := uc.staffRepo.ListByBranch(ctx, branchID, p.Offset, p.Limit)
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

func (uc *merchantUsecase) ListStaffByMerchant(ctx context.Context, merchantID int64, p pagination.Params) ([]domain.StaffMember, *pagination.Metadata, error) {
	data, total, err := uc.staffRepo.ListByMerchant(ctx, merchantID, p.Offset, p.Limit)
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

func (uc *merchantUsecase) UpdateStaff(ctx context.Context, input UpdateStaffParams) (*domain.StaffMember, error) {
	staff, err := uc.staffRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if input.Role != nil {
		staff.Role = *input.Role
	}
	if input.Status != nil {
		staff.Status = *input.Status
	}
	staff.UpdatedAt = time.Now().UTC()

	if err := uc.staffRepo.Update(ctx, staff); err != nil {
		return nil, domain.ErrInternal
	}
	return staff, nil
}

func (uc *merchantUsecase) GetMyStaffAssignments(ctx context.Context, userID, merchantID int64) ([]domain.StaffMember, error) {
	return uc.staffRepo.ListByUserAndMerchant(ctx, userID, merchantID)
}

func (uc *merchantUsecase) SetMyDefaultBranch(ctx context.Context, userID, branchID int64) error {
	_, err := uc.branchRepo.GetByID(ctx, branchID)
	if err != nil {
		return err
	}
	return uc.staffRepo.SetDefaultBranch(ctx, userID, branchID)
}

func (uc *merchantUsecase) RemoveStaff(ctx context.Context, id int64) error {
	return uc.staffRepo.Delete(ctx, id)
}
