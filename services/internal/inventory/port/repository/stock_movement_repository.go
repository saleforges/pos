package repository

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/inventory/domain"
)

type StockMovementRepository interface {
	Create(ctx context.Context, movement *domain.StockMovement) error
	ListByProductItem(ctx context.Context, productItemID int64, merchantID int64) ([]domain.StockMovement, error)
	List(ctx context.Context, merchantID, branchID int64, productItemID *int64, from, to *time.Time) ([]domain.StockMovement, error)
}
