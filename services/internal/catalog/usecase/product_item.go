package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
)

type productItemUsecase struct {
	itemRepo repository.ProductItemRepository
	prodRepo repository.ProductRepository
	unitRepo repository.UnitRepository
}

func NewProductItemUsecase(itemRepo repository.ProductItemRepository, prodRepo repository.ProductRepository, unitRepo repository.UnitRepository) ProductItemUsecase {
	return &productItemUsecase{itemRepo: itemRepo, prodRepo: prodRepo, unitRepo: unitRepo}
}

func (uc *productItemUsecase) Create(ctx context.Context, params CreateProductItemParams) (*domain.ProductItem, error) {
	if params.Name == "" {
		return nil, domain.ErrInvalidProductItem
	}

	// Verify product exists and belongs to the merchant
	product, err := uc.prodRepo.GetByID(ctx, params.ProductID, params.MerchantID)
	if err != nil {
		return nil, err
	}

	if params.UnitID != nil {
		if _, err := uc.unitRepo.GetByID(ctx, *params.UnitID); err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC()
	currency := params.PriceCurrency
	if currency == "" {
		currency = "IDR"
	}
	item := &domain.ProductItem{
		ProductID:      params.ProductID,
		MerchantID:     product.MerchantID,
		Name:           params.Name,
		SKU:            params.SKU,
		UnitID:         params.UnitID,
		Price:          domain.Price{Amount: params.PriceAmount, Currency: currency},
		TrackInventory: params.TrackInventory,
		ImageURL:       params.ImageURL,
		Status:         domain.ProductItemStatusActive,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := uc.itemRepo.Create(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (uc *productItemUsecase) GetByID(ctx context.Context, id int64, merchantID int64) (*domain.ProductItem, error) {
	return uc.itemRepo.GetByID(ctx, id, merchantID)
}

func (uc *productItemUsecase) ListByProduct(ctx context.Context, productID int64, merchantID int64) ([]domain.ProductItem, error) {
	if _, err := uc.prodRepo.GetByID(ctx, productID, merchantID); err != nil {
		return nil, err
	}
	return uc.itemRepo.ListByProduct(ctx, productID, merchantID)
}

func (uc *productItemUsecase) ListByMerchant(ctx context.Context, merchantID int64) ([]domain.ProductItem, error) {
	return uc.itemRepo.ListByMerchant(ctx, merchantID)
}

func (uc *productItemUsecase) Update(ctx context.Context, params UpdateProductItemParams) (*domain.ProductItem, error) {
	item, err := uc.itemRepo.GetByID(ctx, params.ID, params.MerchantID)
	if err != nil {
		return nil, err
	}

	if params.Name != nil {
		if *params.Name == "" {
			return nil, domain.ErrInvalidProductItem
		}
		item.Name = *params.Name
	}
	if params.SKU != nil {
		item.SKU = *params.SKU
	}
	if params.UnitID != nil {
		if _, err := uc.unitRepo.GetByID(ctx, *params.UnitID); err != nil {
			return nil, err
		}
		item.UnitID = params.UnitID
	}
	if params.PriceAmount != nil {
		item.Price.Amount = *params.PriceAmount
	}
	if params.PriceCurrency != nil {
		item.Price.Currency = *params.PriceCurrency
	}
	if params.TrackInventory != nil {
		item.TrackInventory = *params.TrackInventory
	}
	if params.ImageURL != nil {
		item.ImageURL = *params.ImageURL
	}
	if params.Status != nil {
		item.Status = *params.Status
	}
	item.UpdatedAt = time.Now().UTC()

	if err := uc.itemRepo.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (uc *productItemUsecase) Delete(ctx context.Context, id int64, merchantID int64) error {
	if _, err := uc.itemRepo.GetByID(ctx, id, merchantID); err != nil {
		return err
	}
	return uc.itemRepo.Delete(ctx, id, merchantID)
}

func (uc *productItemUsecase) Restore(ctx context.Context, id int64, merchantID int64) (*domain.ProductItem, error) {
	return uc.itemRepo.Restore(ctx, id, merchantID)
}

var _ ProductItemUsecase = (*productItemUsecase)(nil)
