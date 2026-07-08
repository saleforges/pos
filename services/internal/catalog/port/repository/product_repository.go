package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, id string) (*domain.Product, error)
	GetBySKU(ctx context.Context, sku string, merchantID string) (*domain.Product, error)
	List(ctx context.Context, merchantID string, search string, offset, limit int) ([]domain.Product, error)
	ListByCategory(ctx context.Context, categoryID string, offset, limit int) ([]domain.Product, error)
	Count(ctx context.Context, merchantID string, search string) (int, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id string) error
}
