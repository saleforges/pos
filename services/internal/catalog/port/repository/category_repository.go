package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, id string) (*domain.Category, error)
	List(ctx context.Context, merchantID string, offset, limit int) ([]domain.Category, error)
	Update(ctx context.Context, category *domain.Category) error
	Delete(ctx context.Context, id string) error
}
