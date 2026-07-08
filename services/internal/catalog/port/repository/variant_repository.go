package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type VariantRepository interface {
	Create(ctx context.Context, variant *domain.Variant) error
	GetByID(ctx context.Context, id string) (*domain.Variant, error)
	ListByProduct(ctx context.Context, productID string) ([]domain.Variant, error)
	Update(ctx context.Context, variant *domain.Variant) error
	Delete(ctx context.Context, id string) error
}
