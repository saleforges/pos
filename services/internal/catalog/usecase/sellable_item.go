package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
)

type sellableItemUsecase struct {
	itemRepo repository.SellableItemRepository
	prodRepo repository.ProductRepository
	unitRepo repository.UnitRepository
}

func NewSellableItemUsecase(itemRepo repository.SellableItemRepository, prodRepo repository.ProductRepository, unitRepo repository.UnitRepository) SellableItemUsecase {
	return &sellableItemUsecase{itemRepo: itemRepo, prodRepo: prodRepo, unitRepo: unitRepo}
}

func (uc *sellableItemUsecase) Create(ctx context.Context, params CreateSellableItemParams) (*domain.SellableItem, error) {
	if params.Name == "" || params.UnitID == 0 {
		return nil, domain.ErrInvalidSellableItem
	}

	if _, err := uc.prodRepo.GetByID(ctx, params.ProductID); err != nil {
		return nil, err
	}
	if _, err := uc.unitRepo.GetByID(ctx, params.UnitID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	item := &domain.SellableItem{
		ProductID:      params.ProductID,
		Name:           params.Name,
		UnitID:         params.UnitID,
		Price:          params.Price,
		TrackInventory: params.TrackInventory,
		ImageURL:       params.ImageURL,
		Status:         domain.SellableItemStatusActive,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := uc.itemRepo.Create(ctx, item); err != nil {
		return nil, domain.ErrInternal
	}
	return item, nil
}

func (uc *sellableItemUsecase) ListByProduct(ctx context.Context, productID int64) ([]domain.SellableItem, error) {
	if _, err := uc.prodRepo.GetByID(ctx, productID); err != nil {
		return nil, err
	}
	return uc.itemRepo.ListByProduct(ctx, productID)
}

func (uc *sellableItemUsecase) Update(ctx context.Context, params UpdateSellableItemParams) (*domain.SellableItem, error) {
	item, err := uc.itemRepo.GetByID(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	if params.Name != nil {
		if *params.Name == "" {
			return nil, domain.ErrInvalidSellableItem
		}
		item.Name = *params.Name
	}
	if params.UnitID != nil {
		if _, err := uc.unitRepo.GetByID(ctx, *params.UnitID); err != nil {
			return nil, err
		}
		item.UnitID = *params.UnitID
	}
	if params.Price != nil {
		item.Price = *params.Price
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
		return nil, domain.ErrInternal
	}
	return item, nil
}

func (uc *sellableItemUsecase) Delete(ctx context.Context, id int64) error {
	if _, err := uc.itemRepo.GetByID(ctx, id); err != nil {
		return err
	}
	return uc.itemRepo.Delete(ctx, id)
}

func (uc *sellableItemUsecase) Restore(ctx context.Context, id int64) (*domain.SellableItem, error) {
	return uc.itemRepo.Restore(ctx, id)
}

var _ SellableItemUsecase = (*sellableItemUsecase)(nil)
