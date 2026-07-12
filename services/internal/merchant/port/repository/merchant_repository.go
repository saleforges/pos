package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/merchant/domain"
)

type MerchantRepository interface {
	Create(ctx context.Context, merchant *domain.Merchant) error
	GetByID(ctx context.Context, id int64) (*domain.Merchant, error)
	List(ctx context.Context, offset, limit int) ([]domain.Merchant, error)
	Update(ctx context.Context, merchant *domain.Merchant) error
	Delete(ctx context.Context, id int64) error
}
