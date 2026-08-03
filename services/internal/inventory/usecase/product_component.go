package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/inventory/domain"
)

type ProductComponentUsecase interface {
	Create(ctx context.Context, params CreateProductComponentParams) (*domain.ProductComponent, error)
	GetByProductItem(ctx context.Context, productItemID int64, merchantID int64) (*domain.ProductComponent, error)
	List(ctx context.Context, merchantID int64) ([]domain.ProductComponent, error)
	Update(ctx context.Context, params UpdateProductComponentParams) (*domain.ProductComponent, error)
	Delete(ctx context.Context, id int64, merchantID int64) error
}

type CreateProductComponentParams struct {
	MerchantID    int64
	ProductItemID int64
	Items         []CreateProductComponentItemParams
}

type CreateProductComponentItemParams struct {
	ComponentProductItemID int64
	Quantity               float64
	UnitID                 int64
}

type UpdateProductComponentParams struct {
	MerchantID    int64
	ProductItemID int64
	Items         []CreateProductComponentItemParams
}
