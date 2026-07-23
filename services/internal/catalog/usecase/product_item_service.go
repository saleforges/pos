package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type ProductItemUsecase interface {
	Create(ctx context.Context, params CreateProductItemParams) (*domain.ProductItem, error)
	GetByID(ctx context.Context, id int64, merchantID int64) (*domain.ProductItem, error)
	ListByProduct(ctx context.Context, productID int64, merchantID int64) ([]domain.ProductItem, error)
	ListByMerchant(ctx context.Context, merchantID int64) ([]domain.ProductItem, error)
	Update(ctx context.Context, params UpdateProductItemParams) (*domain.ProductItem, error)
	Delete(ctx context.Context, id int64, merchantID int64) error
	Restore(ctx context.Context, id int64, merchantID int64) (*domain.ProductItem, error)
}

type CreateProductItemParams struct {
	MerchantID     int64
	ProductID      int64
	Name           string
	SKU            string
	UnitID         *int64
	PriceAmount    float64
	PriceCurrency  string
	TrackInventory bool
	ImageURL       string
}

type UpdateProductItemParams struct {
	ID             int64
	MerchantID     int64
	Name           *string
	SKU            *string
	UnitID         *int64
	PriceAmount    *float64
	PriceCurrency  *string
	TrackInventory *bool
	ImageURL       *string
	Status         *domain.ProductItemStatus
}
