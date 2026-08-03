package repository

import (
	"context"
)

// StockAdjustmentItem is one line of a stock deduction/restore request sent
// to the Inventory service.
type StockAdjustmentItem struct {
	ProductItemID int64
	Quantity      int64
}

// InventoryClient talks to the Inventory service internal API for stock
// adjustments triggered by the order lifecycle.
type InventoryClient interface {
	DeductStock(ctx context.Context, merchantID, branchID int64, referenceType string, referenceID int64, items []StockAdjustmentItem) error
	RestoreStock(ctx context.Context, merchantID, branchID int64, referenceType string, referenceID int64, items []StockAdjustmentItem) error
}
