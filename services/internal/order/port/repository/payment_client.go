package repository

import (
	"context"
)

// PaymentClient talks to the payment service internal API to create
// gateway payment links for orders.
type PaymentClient interface {
	CreatePayment(ctx context.Context, params CreatePaymentParams) (*PaymentResult, error)
}

// CreatePaymentItem is one order line sent to the payment service.
type CreatePaymentItem struct {
	ItemName  string
	Quantity  float64
	UnitPrice float64
}

type CreatePaymentParams struct {
	MerchantID int64
	OrderID    int64
	Amount     float64
	BuyerName  string
	BuyerEmail string
	BuyerPhone string
	Items      []CreatePaymentItem
}

type PaymentResult struct {
	OrderID    int64  `json:"orderId"`
	PaymentURL string `json:"paymentUrl"`
	SessionID  string `json:"sessionId,omitempty"`
}
