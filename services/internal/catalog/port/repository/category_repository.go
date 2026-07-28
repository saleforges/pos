package repository

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Category, error)
	ListByMerchant(ctx context.Context, merchantID int64) ([]domain.Category, error)
	Update(ctx context.Context, category *domain.Category) error
	Delete(ctx context.Context, id int64, merchantID int64) error
	Restore(ctx context.Context, id int64, merchantID int64) (*domain.Category, error)
	ListUpdatedAfter(ctx context.Context, merchantID int64, after time.Time) ([]domain.Category, error)
}
