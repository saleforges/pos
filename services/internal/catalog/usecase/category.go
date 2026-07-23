package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
)

type categoryUsecase struct {
	catRepo repository.CategoryRepository
}

func NewCategoryUsecase(catRepo repository.CategoryRepository) CategoryUsecase {
	return &categoryUsecase{catRepo: catRepo}
}

func (uc *categoryUsecase) Create(ctx context.Context, params CreateCategoryParams) (*domain.Category, error) {
	if params.Name == "" {
		return nil, domain.ErrInvalidCategory
	}
	if params.ParentID != nil {
		if _, err := uc.catRepo.GetByID(ctx, *params.ParentID); err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC()
	category := &domain.Category{
		MerchantID: params.MerchantID,
		Name:       params.Name,
		ParentID:   params.ParentID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := uc.catRepo.Create(ctx, category); err != nil {
		return nil, domain.ErrInternal
	}
	return category, nil
}

func (uc *categoryUsecase) GetByID(ctx context.Context, id int64) (*domain.Category, error) {
	return uc.catRepo.GetByID(ctx, id)
}

func (uc *categoryUsecase) ListByMerchant(ctx context.Context, merchantID int64) ([]domain.Category, error) {
	return uc.catRepo.ListByMerchant(ctx, merchantID)
}

func (uc *categoryUsecase) Update(ctx context.Context, params UpdateCategoryParams) (*domain.Category, error) {
	category, err := uc.catRepo.GetByID(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	if params.Name != nil {
		if *params.Name == "" {
			return nil, domain.ErrInvalidCategory
		}
		category.Name = *params.Name
	}
	if params.ParentID != nil {
		if _, err := uc.catRepo.GetByID(ctx, *params.ParentID); err != nil {
			return nil, err
		}
		category.ParentID = params.ParentID
	}
	category.UpdatedAt = time.Now().UTC()

	if err := uc.catRepo.Update(ctx, category); err != nil {
		return nil, domain.ErrInternal
	}
	return category, nil
}

func (uc *categoryUsecase) Delete(ctx context.Context, id int64) error {
	if _, err := uc.catRepo.GetByID(ctx, id); err != nil {
		return err
	}
	return uc.catRepo.Delete(ctx, id)
}

func (uc *categoryUsecase) Restore(ctx context.Context, id int64) (*domain.Category, error) {
	return uc.catRepo.Restore(ctx, id)
}

var _ CategoryUsecase = (*categoryUsecase)(nil)
