package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/inventory/domain"
)

type ProductComponentRepository interface {
	Create(ctx context.Context, component *domain.ProductComponent) error
	GetByProductItem(ctx context.Context, productItemID int64, merchantID int64) (*domain.ProductComponent, error)
	List(ctx context.Context, merchantID int64) ([]domain.ProductComponent, error)
	Update(ctx context.Context, component *domain.ProductComponent) error
	Delete(ctx context.Context, id int64, merchantID int64) error
}
