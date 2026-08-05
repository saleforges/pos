package repository

import (
	"context"
)

// OrderItem is one line of an order, needed to build the gateway request.
type OrderItem struct {
	ItemName  string
	Quantity  float64
	UnitPrice float64
}

// OrderInfo is the order snapshot the payment service needs to validate
// and size a gateway payment.
type OrderInfo struct {
	ID         int64
	MerchantID int64
	BranchID   int64
	Status     string
	Total      float64
	PaidAmount float64
	Items      []OrderItem
}

// OrderClient talks to the order service internal API.
type OrderClient interface {
	// GetOrder fetches the order snapshot used to validate a payment.
	GetOrder(ctx context.Context, orderID, merchantID int64) (*OrderInfo, error)
	// NotifyPaid records a gateway-confirmed payment on an order.
	NotifyPaid(ctx context.Context, orderID, merchantID int64, amount float64, method string) error
}
