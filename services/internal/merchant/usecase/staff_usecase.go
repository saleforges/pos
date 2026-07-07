package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/pkg/id"
	"github.com/saleforge/pos/services/internal/merchant/domain"
)

type StaffUsecase interface {
	AssignStaff(ctx context.Context, input AssignStaffInput) (*domain.StaffMember, error)
	GetStaff(ctx context.Context, id string) (*domain.StaffMember, error)
	ListStaffByBranch(ctx context.Context, branchID string) ([]domain.StaffMember, error)
	ListStaffByMerchant(ctx context.Context, merchantID string) ([]domain.StaffMember, error)
	GetMyStaffAssignments(ctx context.Context, userID, merchantID string) ([]domain.StaffMember, error)
	SetMyDefaultBranch(ctx context.Context, userID, branchID string) error
	UpdateStaff(ctx context.Context, input UpdateStaffInput) (*domain.StaffMember, error)
	RemoveStaff(ctx context.Context, id string) error
}

type AssignStaffInput struct {
	MerchantID string
	BranchID   string
	UserID     string
	Role       domain.StaffRole
	IsDefault  bool
}

type UpdateStaffInput struct {
	ID     string
	Role   *domain.StaffRole
	Status *domain.StaffStatus
}

func (uc *merchantUsecase) AssignStaff(ctx context.Context, input AssignStaffInput) (*domain.StaffMember, error) {
	if input.MerchantID == "" || input.BranchID == "" || input.UserID == "" || input.Role == "" {
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

	existing, _ := uc.staffRepo.GetByUserAndBranch(ctx, input.UserID, input.BranchID)
	if existing != nil {
		return nil, domain.ErrStaffExists
	}

	now := time.Now().UTC()
	staff := &domain.StaffMember{
		ID:         id.Generate(),
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
		_ = uc.staffRepo.SetDefaultBranch(ctx, input.UserID, staff.ID)
	}

	if err := uc.staffRepo.Create(ctx, staff); err != nil {
		return nil, domain.ErrInternal
	}
	return staff, nil
}

func (uc *merchantUsecase) GetStaff(ctx context.Context, id string) (*domain.StaffMember, error) {
	return uc.staffRepo.GetByID(ctx, id)
}

func (uc *merchantUsecase) ListStaffByBranch(ctx context.Context, branchID string) ([]domain.StaffMember, error) {
	return uc.staffRepo.ListByBranch(ctx, branchID)
}

func (uc *merchantUsecase) ListStaffByMerchant(ctx context.Context, merchantID string) ([]domain.StaffMember, error) {
	return uc.staffRepo.ListByMerchant(ctx, merchantID)
}

func (uc *merchantUsecase) UpdateStaff(ctx context.Context, input UpdateStaffInput) (*domain.StaffMember, error) {
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

func (uc *merchantUsecase) GetMyStaffAssignments(ctx context.Context, userID, merchantID string) ([]domain.StaffMember, error) {
	return uc.staffRepo.ListByUserAndMerchant(ctx, userID, merchantID)
}

func (uc *merchantUsecase) SetMyDefaultBranch(ctx context.Context, userID, branchID string) error {
	_, err := uc.branchRepo.GetByID(ctx, branchID)
	if err != nil {
		return err
	}
	return uc.staffRepo.SetDefaultBranch(ctx, userID, branchID)
}

func (uc *merchantUsecase) RemoveStaff(ctx context.Context, id string) error {
	return uc.staffRepo.Delete(ctx, id)
}
