package domain

import "time"

type MovementType string

const (
	MovementTypeStockIn     MovementType = "stock_in"
	MovementTypeStockOut    MovementType = "stock_out"
	MovementTypeAdjustment  MovementType = "adjustment"
)

type StockMovement struct {
	ID            int64        `json:"id"`
	MerchantID    int64        `json:"merchantId"`
	BranchID      int64        `json:"branchId"`
	ProductItemID int64        `json:"productItemId"`
	Type          MovementType `json:"type"`
	Quantity      int64        `json:"quantity"`
	ReferenceType string       `json:"referenceType,omitempty"`
	ReferenceID   *int64       `json:"referenceId,omitempty"`
	Note          string       `json:"note,omitempty"`
	CreatedAt     time.Time    `json:"createdAt"`
}
