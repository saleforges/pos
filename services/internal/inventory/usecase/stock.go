package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/inventory/domain"
)

type StockUsecase interface {
	Create(ctx context.Context, params CreateStockParams) (*domain.Stock, error)
	GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Stock, error)
	List(ctx context.Context, merchantID int64) ([]domain.Stock, error)
	Update(ctx context.Context, params UpdateStockParams) (*domain.Stock, error)
}

type CreateStockParams struct {
	MerchantID    int64
	BranchID      int64
	ProductItemID int64
	Available     int64
}

type UpdateStockParams struct {
	ID         int64
	MerchantID int64
	Available  int64
}
