package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/pkg/id"
	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/internal/merchant/port/repository"
)

type MerchantUsecase interface {
	CreateMerchant(ctx context.Context, input CreateMerchantInput) (*domain.Merchant, error)
	GetMerchant(ctx context.Context, id string) (*domain.Merchant, error)
	ListMerchants(ctx context.Context, offset, limit int) ([]domain.Merchant, error)
	UpdateMerchant(ctx context.Context, input UpdateMerchantInput) (*domain.Merchant, error)
	DeleteMerchant(ctx context.Context, id string) error
}

type merchantUsecase struct {
	merchantRepo repository.MerchantRepository
	branchRepo   repository.BranchRepository
	staffRepo    repository.StaffRepository
}

func NewMerchantUsecase(merchantRepo repository.MerchantRepository, branchRepo repository.BranchRepository, staffRepo repository.StaffRepository) *merchantUsecase {
	return &merchantUsecase{merchantRepo: merchantRepo, branchRepo: branchRepo, staffRepo: staffRepo}
}

type CreateMerchantInput struct {
	Name      string
	LegalName string
	Address   string
	Phone     string
	Email     string
	TaxID     string
	Settings  domain.MerchantSettings
}

type UpdateMerchantInput struct {
	ID        string
	Name      *string
	LegalName *string
	Address   *string
	Phone     *string
	Email     *string
	TaxID     *string
	Status    *domain.MerchantStatus
	Settings  *domain.MerchantSettings
}

func (uc *merchantUsecase) CreateMerchant(ctx context.Context, input CreateMerchantInput) (*domain.Merchant, error) {
	if input.Name == "" || input.Email == "" {
		return nil, domain.ErrInvalidMerchant
	}

	now := time.Now().UTC()
	merchant := &domain.Merchant{
		ID:        id.Generate(),
		Name:      input.Name,
		LegalName: input.LegalName,
		Address:   input.Address,
		Phone:     input.Phone,
		Email:     input.Email,
		TaxID:     input.TaxID,
		Status:    domain.MerchantStatusActive,
		Settings:  input.Settings,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if merchant.Settings.Currency == "" {
		merchant.Settings.Currency = "IDR"
	}
	if merchant.Settings.Timezone == "" {
		merchant.Settings.Timezone = "Asia/Jakarta"
	}
	if merchant.Settings.LowStockThreshold == 0 {
		merchant.Settings.LowStockThreshold = 10
	}

	if err := uc.merchantRepo.Create(ctx, merchant); err != nil {
		return nil, domain.ErrInternal
	}
	return merchant, nil
}

func (uc *merchantUsecase) GetMerchant(ctx context.Context, id string) (*domain.Merchant, error) {
	return uc.merchantRepo.GetByID(ctx, id)
}

func (uc *merchantUsecase) ListMerchants(ctx context.Context, offset, limit int) ([]domain.Merchant, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return uc.merchantRepo.List(ctx, offset, limit)
}

func (uc *merchantUsecase) UpdateMerchant(ctx context.Context, input UpdateMerchantInput) (*domain.Merchant, error) {
	merchant, err := uc.merchantRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		merchant.Name = *input.Name
	}
	if input.LegalName != nil {
		merchant.LegalName = *input.LegalName
	}
	if input.Address != nil {
		merchant.Address = *input.Address
	}
	if input.Phone != nil {
		merchant.Phone = *input.Phone
	}
	if input.Email != nil {
		merchant.Email = *input.Email
	}
	if input.TaxID != nil {
		merchant.TaxID = *input.TaxID
	}
	if input.Status != nil {
		merchant.Status = *input.Status
	}
	if input.Settings != nil {
		merchant.Settings = *input.Settings
	}
	merchant.UpdatedAt = time.Now().UTC()

	if err := uc.merchantRepo.Update(ctx, merchant); err != nil {
		return nil, domain.ErrInternal
	}
	return merchant, nil
}

func (uc *merchantUsecase) DeleteMerchant(ctx context.Context, id string) error {
	return uc.merchantRepo.Delete(ctx, id)
}
