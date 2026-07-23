package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type SellableItemRepository interface {
	Create(ctx context.Context, item *domain.SellableItem) error
	GetByID(ctx context.Context, id int64) (*domain.SellableItem, error)
	ListByProduct(ctx context.Context, productID int64) ([]domain.SellableItem, error)
	Update(ctx context.Context, item *domain.SellableItem) error
	Delete(ctx context.Context, id int64) error
	Restore(ctx context.Context, id int64) (*domain.SellableItem, error)
}

type BarcodeRepository interface {
	Create(ctx context.Context, barcode *domain.SellableItemBarcode) error
	GetByBarcode(ctx context.Context, barcode string) (*domain.SellableItemBarcode, error)
	ListBySellableItem(ctx context.Context, sellableItemID int64) ([]domain.SellableItemBarcode, error)
	Delete(ctx context.Context, id int64) error
}
