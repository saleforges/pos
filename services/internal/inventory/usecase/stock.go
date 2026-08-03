package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/inventory/domain"
)

type StockUsecase interface {
	Create(ctx context.Context, params CreateStockParams) (*domain.Stock, error)
	GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Stock, error)
	List(ctx context.Context, merchantID int64) ([]domain.Stock, error)
	Update(ctx context.Context, params UpdateStockParams) (*domain.Stock, error)
	Deduct(ctx context.Context, params AdjustStockParams) error
	Restore(ctx context.Context, params AdjustStockParams) error
	Sync(ctx context.Context, merchantID int64, lastSync *time.Time) (*StockSyncResult, error)
}

type StockSyncResult struct {
	Stocks    []domain.Stock `json:"stocks"`
	SyncToken string         `json:"syncToken"`
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

// AdjustStockParams is a batch stock change with movement provenance
// (reference_type/reference_id, e.g. order/42). Used by the order flow.
type AdjustStockParams struct {
	MerchantID    int64
	BranchID      int64
	ReferenceType string
	ReferenceID   int64
	Items         []AdjustStockItem
}

type AdjustStockItem struct {
	ProductItemID int64
	Quantity      int64
}
