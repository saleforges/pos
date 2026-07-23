package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type ProductItemRepository interface {
	Create(ctx context.Context, item *domain.ProductItem) error
	GetByID(ctx context.Context, id int64, merchantID int64) (*domain.ProductItem, error)
	ListByProduct(ctx context.Context, productID int64, merchantID int64) ([]domain.ProductItem, error)
	ListByMerchant(ctx context.Context, merchantID int64) ([]domain.ProductItem, error)
	Update(ctx context.Context, item *domain.ProductItem) error
	Delete(ctx context.Context, id int64, merchantID int64) error
	Restore(ctx context.Context, id int64, merchantID int64) (*domain.ProductItem, error)
}

type ProductItemBarcodeRepository interface {
	Create(ctx context.Context, barcode *domain.ProductItemBarcode) error
	GetByBarcode(ctx context.Context, barcode string) (*domain.ProductItemBarcode, error)
	ListByProductItem(ctx context.Context, productItemID int64) ([]domain.ProductItemBarcode, error)
	Delete(ctx context.Context, id int64) error
}
