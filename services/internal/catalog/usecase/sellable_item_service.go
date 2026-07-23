package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type SellableItemUsecase interface {
	Create(ctx context.Context, params CreateSellableItemParams) (*domain.SellableItem, error)
	ListByProduct(ctx context.Context, productID int64) ([]domain.SellableItem, error)
	Update(ctx context.Context, params UpdateSellableItemParams) (*domain.SellableItem, error)
	Delete(ctx context.Context, id int64) error
	Restore(ctx context.Context, id int64) (*domain.SellableItem, error)
}

type CreateSellableItemParams struct {
	ProductID      int64
	Name           string
	UnitID         int64
	Price          float64
	TrackInventory bool
	ImageURL       string
}

type UpdateSellableItemParams struct {
	ID             int64
	Name           *string
	UnitID         *int64
	Price          *float64
	TrackInventory *bool
	ImageURL       *string
	Status         *domain.SellableItemStatus
}
