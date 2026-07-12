package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
	"github.com/saleforge/pos/services/pkg/logger"
	"github.com/saleforge/pos/services/pkg/otel"
)

type variantUsecase struct {
	varRepo  repository.VariantRepository
	prodRepo repository.ProductRepository
}

func NewVariantUsecase(varRepo repository.VariantRepository, prodRepo repository.ProductRepository) VariantUsecase {
	return &variantUsecase{varRepo: varRepo, prodRepo: prodRepo}
}

func (uc *variantUsecase) Create(ctx context.Context, input CreateVariantInput) (*domain.Variant, error) {
	ctx, span := otel.StartSpan(ctx, "variant.Create")
	defer span.End()

	if _, err := uc.prodRepo.GetByID(ctx, input.ProductID); err != nil {
		return nil, err
	}

	v := &domain.Variant{
		ProductID: input.ProductID,
		Name:      input.Name,
		SKU:       input.SKU,
		Barcode:   input.Barcode,
		Price:     input.Price,
		Cost:      input.Cost,
		ImageURL:  input.ImageURL,
		SortOrder: input.SortOrder,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if err := uc.varRepo.Create(ctx, v); err != nil {
		logger.Error("variant.Create: create failed", "error", err.Error())
		return nil, domain.ErrInternal
	}
	return v, nil
}

func (uc *variantUsecase) ListByProduct(ctx context.Context, productID int64) ([]domain.Variant, error) {
	ctx, span := otel.StartSpan(ctx, "variant.ListByProduct")
	defer span.End()

	if _, err := uc.prodRepo.GetByID(ctx, productID); err != nil {
		return nil, err
	}
	return uc.varRepo.ListByProduct(ctx, productID)
}

func (uc *variantUsecase) Update(ctx context.Context, input UpdateVariantInput) (*domain.Variant, error) {
	ctx, span := otel.StartSpan(ctx, "variant.Update")
	defer span.End()

	v, err := uc.varRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		v.Name = *input.Name
	}
	if input.SKU != nil {
		v.SKU = *input.SKU
	}
	if input.Barcode != nil {
		v.Barcode = *input.Barcode
	}
	if input.Price != nil {
		v.Price = *input.Price
	}
	if input.Cost != nil {
		v.Cost = *input.Cost
	}
	if input.ImageURL != nil {
		v.ImageURL = *input.ImageURL
	}
	if input.SortOrder != nil {
		v.SortOrder = *input.SortOrder
	}
	v.UpdatedAt = time.Now().UTC()

	if err := uc.varRepo.Update(ctx, v); err != nil {
		logger.Error("variant.Update: update failed", "error", err.Error())
		return nil, domain.ErrInternal
	}
	return v, nil
}

func (uc *variantUsecase) Delete(ctx context.Context, id int64) error {
	ctx, span := otel.StartSpan(ctx, "variant.Delete")
	defer span.End()

	_, err := uc.varRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	return uc.varRepo.Delete(ctx, id)
}
