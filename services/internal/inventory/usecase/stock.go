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
	Produce(ctx context.Context, params ProduceParams) (*domain.Stock, error)
	Opname(ctx context.Context, params OpnameParams) (*domain.Stock, error)
	Transfer(ctx context.Context, params TransferStockParams) error
	Sync(ctx context.Context, merchantID, branchID int64, lastSync *time.Time) (*StockSyncResult, error)
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

type AdjustStockParams struct {
	MerchantID    int64
	BranchID      int64
	ReferenceType string
	ReferenceID   int64
	Items         []AdjustStockItem
}

type AdjustStockItem struct {
	ProductItemID int64 `json:"productItemId"`
	Quantity      int64 `json:"quantity"`
}

type ProduceParams struct {
	MerchantID    int64
	BranchID      int64
	ProductItemID int64
	Quantity      int64
}

type OpnameParams struct {
	MerchantID     int64
	BranchID       int64
	ProductItemID  int64
	ActualQuantity int64
	Reason         string
}

type TransferStockParams struct {
	MerchantID   int64
	FromBranchID int64
	ToBranchID   int64
	Items        []AdjustStockItem
}
