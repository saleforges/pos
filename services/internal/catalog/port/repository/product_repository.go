package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, id int64) (*domain.Product, error)
	List(ctx context.Context, merchantID int64, search string, offset, limit int) ([]domain.Product, int, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id int64) error
	Restore(ctx context.Context, id int64) (*domain.Product, error)
}
