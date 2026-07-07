package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/pkg/id"
	"github.com/saleforge/pos/services/internal/merchant/domain"
)

type BranchUsecase interface {
	CreateBranch(ctx context.Context, input CreateBranchInput) (*domain.Branch, error)
	GetBranch(ctx context.Context, id string) (*domain.Branch, error)
	ListBranches(ctx context.Context, merchantID string) ([]domain.Branch, error)
	UpdateBranch(ctx context.Context, input UpdateBranchInput) (*domain.Branch, error)
	DeleteBranch(ctx context.Context, id string) error
}

type CreateBranchInput struct {
	MerchantID     string
	Name           string
	Code           string
	Address        string
	Phone          string
	OperatingDays  []string
	OperatingHours *domain.OperatingHours
}

type UpdateBranchInput struct {
	ID             string
	Name           *string
	Address        *string
	Phone          *string
	Status         *domain.BranchStatus
	OperatingDays  []string
	OperatingHours *domain.OperatingHours
}

func (uc *merchantUsecase) CreateBranch(ctx context.Context, input CreateBranchInput) (*domain.Branch, error) {
	if input.Name == "" || input.MerchantID == "" || input.Code == "" {
		return nil, domain.ErrInvalidBranch
	}

	_, err := uc.merchantRepo.GetByID(ctx, input.MerchantID)
	if err != nil {
		return nil, domain.ErrMerchantNotFound
	}

	now := time.Now().UTC()
	branch := &domain.Branch{
		ID:             id.Generate(),
		MerchantID:     input.MerchantID,
		Name:           input.Name,
		Code:           input.Code,
		Address:        input.Address,
		Phone:          input.Phone,
		Status:         domain.BranchStatusActive,
		OperatingDays:  input.OperatingDays,
		OperatingHours: input.OperatingHours,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := uc.branchRepo.Create(ctx, branch); err != nil {
		return nil, domain.ErrInternal
	}
	return branch, nil
}

func (uc *merchantUsecase) GetBranch(ctx context.Context, id string) (*domain.Branch, error) {
	return uc.branchRepo.GetByID(ctx, id)
}

func (uc *merchantUsecase) ListBranches(ctx context.Context, merchantID string) ([]domain.Branch, error) {
	return uc.branchRepo.ListByMerchant(ctx, merchantID)
}

func (uc *merchantUsecase) UpdateBranch(ctx context.Context, input UpdateBranchInput) (*domain.Branch, error) {
	branch, err := uc.branchRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		branch.Name = *input.Name
	}
	if input.Address != nil {
		branch.Address = *input.Address
	}
	if input.Phone != nil {
		branch.Phone = *input.Phone
	}
	if input.Status != nil {
		branch.Status = *input.Status
	}
	if input.OperatingDays != nil {
		branch.OperatingDays = input.OperatingDays
	}
	if input.OperatingHours != nil {
		branch.OperatingHours = input.OperatingHours
	}
	branch.UpdatedAt = time.Now().UTC()

	if err := uc.branchRepo.Update(ctx, branch); err != nil {
		return nil, domain.ErrInternal
	}
	return branch, nil
}

func (uc *merchantUsecase) DeleteBranch(ctx context.Context, id string) error {
	return uc.branchRepo.Delete(ctx, id)
}
