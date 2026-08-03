package repository

import (
	"context"
)

// CreatePaymentItem is one order line sent to the payment gateway.
type CreatePaymentItem struct {
	ItemName  string
	Quantity  float64
	UnitPrice float64
}

// CreatePaymentParams is the order summary sent to the gateway.
type CreatePaymentParams struct {
	ReferenceID string
	BuyerName   string
	BuyerEmail  string
	BuyerPhone  string
	Items       []CreatePaymentItem
}

// PaymentResult is the gateway response with a redirect URL.
type PaymentResult struct {
	SessionID  string
	PaymentURL string
}

// GatewayClient creates payment links via an external gateway (iPaymu).
type GatewayClient interface {
	CreatePayment(ctx context.Context, params CreatePaymentParams) (*PaymentResult, error)
}
