package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
	"github.com/saleforge/pos/services/pkg/logger"
	"github.com/saleforge/pos/services/pkg/otel"
)

type productUsecase struct {
	prodRepo repository.ProductRepository
	catRepo  repository.CategoryRepository
}

func NewProductUsecase(prodRepo repository.ProductRepository, catRepo repository.CategoryRepository) ProductUsecase {
	return &productUsecase{prodRepo: prodRepo, catRepo: catRepo}
}

func (uc *productUsecase) Create(ctx context.Context, input CreateProductInput) (*domain.Product, error) {
	ctx, span := otel.StartSpan(ctx, "product.Create")
	defer span.End()

	if _, err := uc.catRepo.GetByID(ctx, input.CategoryID); err != nil {
		return nil, err
	}
	if existing, _ := uc.prodRepo.GetBySKU(ctx, input.SKU, input.MerchantID); existing != nil {
		return nil, domain.ErrSkuExists
	}

	prod := &domain.Product{
		MerchantID:  input.MerchantID,
		CategoryID:  input.CategoryID,
		Name:        input.Name,
		SKU:         input.SKU,
		Barcode:     input.Barcode,
		Description: input.Description,
		Price:       input.Price,
		Cost:        input.Cost,
		TaxRate:     input.TaxRate,
		Unit:        input.Unit,
		ImageURL:    input.ImageURL,
		Status:      domain.ProductStatusActive,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	if err := uc.prodRepo.Create(ctx, prod); err != nil {
		logger.Error("product.Create: create failed", "error", err.Error())
		return nil, domain.ErrInternal
	}
	return prod, nil
}

func (uc *productUsecase) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	ctx, span := otel.StartSpan(ctx, "product.GetByID")
	defer span.End()

	return uc.prodRepo.GetByID(ctx, id)
}

func (uc *productUsecase) List(ctx context.Context, merchantID int64, search string, offset, limit int) (*PaginatedResult[domain.Product], error) {
	ctx, span := otel.StartSpan(ctx, "product.List")
	defer span.End()

	items, err := uc.prodRepo.List(ctx, merchantID, search, offset, limit)
	if err != nil {
		return nil, err
	}
	total, err := uc.prodRepo.Count(ctx, merchantID, search)
	if err != nil {
		return nil, err
	}
	return &PaginatedResult[domain.Product]{
		Items: items,
		Meta:  PaginationMeta{Total: total, Offset: offset, Limit: limit},
	}, nil
}

func (uc *productUsecase) Update(ctx context.Context, input UpdateProductInput) (*domain.Product, error) {
	ctx, span := otel.StartSpan(ctx, "product.Update")
	defer span.End()

	prod, err := uc.prodRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if input.CategoryID != nil {
		if _, err := uc.catRepo.GetByID(ctx, *input.CategoryID); err != nil {
			return nil, err
		}
		prod.CategoryID = *input.CategoryID
	}
	if input.Name != nil {
		prod.Name = *input.Name
	}
	if input.SKU != nil {
		prod.SKU = *input.SKU
	}
	if input.Barcode != nil {
		prod.Barcode = *input.Barcode
	}
	if input.Description != nil {
		prod.Description = *input.Description
	}
	if input.Price != nil {
		prod.Price = *input.Price
	}
	if input.Cost != nil {
		prod.Cost = *input.Cost
	}
	if input.TaxRate != nil {
		prod.TaxRate = *input.TaxRate
	}
	if input.Unit != nil {
		prod.Unit = *input.Unit
	}
	if input.ImageURL != nil {
		prod.ImageURL = *input.ImageURL
	}
	if input.Status != nil {
		prod.Status = *input.Status
	}
	prod.UpdatedAt = time.Now().UTC()

	if err := uc.prodRepo.Update(ctx, prod); err != nil {
		logger.Error("product.Update: update failed", "error", err.Error())
		return nil, domain.ErrInternal
	}
	return prod, nil
}

func (uc *productUsecase) Delete(ctx context.Context, id int64) error {
	ctx, span := otel.StartSpan(ctx, "product.Delete")
	defer span.End()

	_, err := uc.prodRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return uc.prodRepo.Delete(ctx, id)
}
