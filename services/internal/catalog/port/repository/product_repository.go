package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, id int64) (*domain.Product, error)
	GetBySKU(ctx context.Context, sku string, merchantID int64) (*domain.Product, error)
	List(ctx context.Context, merchantID int64, search string, offset, limit int) ([]domain.Product, error)
	ListByCategory(ctx context.Context, categoryID int64, offset, limit int) ([]domain.Product, error)
	Count(ctx context.Context, merchantID int64, search string) (int, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id int64) error
}
