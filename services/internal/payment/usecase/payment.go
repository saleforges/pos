package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/payment/domain"
)

type PaymentUsecase interface {
	Create(ctx context.Context, params CreatePaymentParams) (*domain.PaymentTransaction, error)
	HandleCallback(ctx context.Context, params CallbackParams) error
	GetByID(ctx context.Context, id int64, merchantID int64) (*domain.PaymentTransaction, error)
	GetByOrderID(ctx context.Context, orderID int64, merchantID int64) (*domain.PaymentTransaction, error)
}

// CreatePaymentParams comes from the order service internal API.
type CreatePaymentParams struct {
	MerchantID int64
	OrderID    int64
	Amount     float64
	BuyerName  string
	BuyerEmail string
	BuyerPhone string
	Items      []CreatePaymentItem
}

type CreatePaymentItem struct {
	ItemName  string
	Quantity  float64
	UnitPrice float64
}

// CallbackParams is the gateway webhook payload.
type CallbackParams struct {
	ReferenceID string
	Status      string
	StatusCode  int
	Amount      string
	GatewayRef  string // trx_id from gateway
	Via         string // va / qris / ew / cod
	Channel     string
	Signature   string
	MerchantVA  string
}
