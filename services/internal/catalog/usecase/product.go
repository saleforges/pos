package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
	"github.com/saleforge/pos/services/pkg/pagination"
)

type productUsecase struct {
	prodRepo repository.ProductRepository
	catRepo  repository.CategoryRepository
	unitRepo repository.UnitRepository
}

func NewProductUsecase(prodRepo repository.ProductRepository, catRepo repository.CategoryRepository, unitRepo repository.UnitRepository) ProductUsecase {
	return &productUsecase{prodRepo: prodRepo, catRepo: catRepo, unitRepo: unitRepo}
}

func (uc *productUsecase) Create(ctx context.Context, params CreateProductParams) (*domain.Product, error) {
	if params.Name == "" {
		return nil, domain.ErrInvalidProduct
	}

	if _, err := uc.catRepo.GetByID(ctx, params.CategoryID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	product := &domain.Product{
		MerchantID:  params.MerchantID,
		CategoryID:  params.CategoryID,
		Name:        params.Name,
		Description: params.Description,
		ImageURL:    params.ImageURL,
		Status:      domain.ProductStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.prodRepo.Create(ctx, product); err != nil {
		return nil, domain.ErrInternal
	}
	return product, nil
}

func (uc *productUsecase) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	return uc.prodRepo.GetByID(ctx, id)
}

func (uc *productUsecase) List(ctx context.Context, merchantID int64, search string, p pagination.Params) ([]domain.Product, *pagination.Metadata, error) {
	data, total, err := uc.prodRepo.List(ctx, merchantID, search, p.Offset, p.Limit)
	if err != nil {
		return nil, nil, err
	}
	meta := &pagination.Metadata{
		Total:       int64(total),
		Offset:      p.Offset,
		Limit:       p.Limit,
		ReturnCount: len(data),
	}
	return data, meta, nil
}

func (uc *productUsecase) Update(ctx context.Context, params UpdateProductParams) (*domain.Product, error) {
	product, err := uc.prodRepo.GetByID(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	if params.CategoryID != nil {
		if _, err := uc.catRepo.GetByID(ctx, *params.CategoryID); err != nil {
			return nil, err
		}
		product.CategoryID = *params.CategoryID
	}
	if params.Name != nil {
		if *params.Name == "" {
			return nil, domain.ErrInvalidProduct
		}
		product.Name = *params.Name
	}
	if params.Description != nil {
		product.Description = *params.Description
	}
	if params.ImageURL != nil {
		product.ImageURL = *params.ImageURL
	}
	if params.Status != nil {
		product.Status = *params.Status
	}
	product.UpdatedAt = time.Now().UTC()

	if err := uc.prodRepo.Update(ctx, product); err != nil {
		return nil, domain.ErrInternal
	}
	return product, nil
}

func (uc *productUsecase) Delete(ctx context.Context, id int64) error {
	if _, err := uc.prodRepo.GetByID(ctx, id); err != nil {
		return err
	}
	return uc.prodRepo.Delete(ctx, id)
}

// compile-time check
var _ ProductUsecase = (*productUsecase)(nil)
