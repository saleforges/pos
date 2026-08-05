package repository

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/inventory/domain"
)

type StockRepository interface {
	Create(ctx context.Context, stock *domain.Stock) error
	GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Stock, error)
	List(ctx context.Context, merchantID int64) ([]domain.Stock, error)
	ListChangedSince(ctx context.Context, merchantID int64, since *time.Time) ([]domain.Stock, error)
	SyncByBranch(ctx context.Context, merchantID, branchID int64, since *time.Time) ([]domain.Stock, error)
	Update(ctx context.Context, stock *domain.Stock) error
}

// StockAdjustmentItem is a single line in a stock deduction/restore batch.
type StockAdjustmentItem struct {
	ProductItemID int64
	Quantity      int64
}

// StockAdjustmentRepository applies batch stock changes atomically and
// records a stock movement per line. Used by the order flow (deduct on sale,
// restore on cancel) via the internal API.
type StockAdjustmentRepository interface {
	Deduct(ctx context.Context, merchantID, branchID int64, referenceType string, referenceID int64, items []StockAdjustmentItem) error
	Restore(ctx context.Context, merchantID, branchID int64, referenceType string, referenceID int64, items []StockAdjustmentItem) error
}
