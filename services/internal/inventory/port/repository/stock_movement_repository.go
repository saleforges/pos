package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/inventory/domain"
)

type StockMovementRepository interface {
	Create(ctx context.Context, movement *domain.StockMovement) error
	ListByProductItem(ctx context.Context, productItemID int64, merchantID int64) ([]domain.StockMovement, error)
}
