package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/pkg/pagination"
)

type CreateBranchParams struct {
	MerchantID     int64
	Name           string
	Code           string
	Address        string
	Phone          string
	OperatingDays  []string
	OperatingHours *domain.OperatingHours
}

type UpdateBranchParams struct {
	ID             int64
	Name           *string
	Address        *string
	Phone          *string
	Status         *domain.BranchStatus
	OperatingDays  []string
	OperatingHours *domain.OperatingHours
}

func (uc *merchantUsecase) CreateBranch(ctx context.Context, input CreateBranchParams) (*domain.Branch, error) {
	if input.Name == "" || input.Code == "" {
		return nil, domain.ErrInvalidBranch
	}

	_, err := uc.merchantRepo.GetByID(ctx, input.MerchantID)
	if err != nil {
		return nil, domain.ErrMerchantNotFound
	}

	now := time.Now().UTC()
	branch := &domain.Branch{
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
	if branch.OperatingDays == nil {
		branch.OperatingDays = []string{}
	}
	if branch.OperatingHours == nil {
		branch.OperatingHours = &domain.OperatingHours{}
	}

	if err := uc.branchRepo.Create(ctx, branch); err != nil {
		if errors.Is(err, domain.ErrBranchExists) {
			return nil, err
		}
		return nil, domain.ErrInternal
	}
	return branch, nil
}

func (uc *merchantUsecase) GetBranch(ctx context.Context, id int64) (*domain.Branch, error) {
	return uc.branchRepo.GetByID(ctx, id)
}

func (uc *merchantUsecase) ListBranches(ctx context.Context, merchantID int64, p pagination.Params) ([]domain.Branch, *pagination.Metadata, error) {
	data, total, err := uc.branchRepo.ListByMerchant(ctx, merchantID, p.Offset, p.Limit)
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

func (uc *merchantUsecase) UpdateBranch(ctx context.Context, input UpdateBranchParams) (*domain.Branch, error) {
	if input.Name != nil && *input.Name == "" {
		return nil, domain.ErrInvalidBranch
	}

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
	if branch.OperatingHours == nil {
		branch.OperatingHours = &domain.OperatingHours{}
	}
	branch.UpdatedAt = time.Now().UTC()

	if err := uc.branchRepo.Update(ctx, branch); err != nil {
		if errors.Is(err, domain.ErrBranchNotFound) {
			return nil, err
		}
		return nil, domain.ErrInternal
	}
	return branch, nil
}

func (uc *merchantUsecase) DeleteBranch(ctx context.Context, id int64) error {
	return uc.branchRepo.Delete(ctx, id)
}
