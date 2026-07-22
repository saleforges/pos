package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/pkg/pagination"
)

// MerchantUsecase defines merchant CRUD operations.
type MerchantUsecase interface {
	CreateMerchant(ctx context.Context, input CreateMerchantParams) (*domain.Merchant, error)
	GetMerchant(ctx context.Context, id int64) (*domain.Merchant, error)
	ListMerchants(ctx context.Context, p pagination.Params) ([]domain.Merchant, *pagination.Metadata, error)
	UpdateMerchant(ctx context.Context, input UpdateMerchantParams) (*domain.Merchant, error)
	DeleteMerchant(ctx context.Context, id int64) error
}
