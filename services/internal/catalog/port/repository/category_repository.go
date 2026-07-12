package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, id int64) (*domain.Category, error)
	List(ctx context.Context, merchantID int64, search string, offset, limit int) ([]domain.Category, error)
	Count(ctx context.Context, merchantID int64, search string) (int, error)
	Update(ctx context.Context, category *domain.Category) error
	Delete(ctx context.Context, id int64) error
}
