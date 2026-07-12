package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type VariantRepository interface {
	Create(ctx context.Context, variant *domain.Variant) error
	GetByID(ctx context.Context, id int64) (*domain.Variant, error)
	ListByProduct(ctx context.Context, productID int64) ([]domain.Variant, error)
	CountByProduct(ctx context.Context, productID int64) (int, error)
	Update(ctx context.Context, variant *domain.Variant) error
	Delete(ctx context.Context, id int64) error
}
