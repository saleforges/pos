package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/payment/domain"
)

type PaymentUsecase interface {
	Create(ctx context.Context, params CreatePaymentParams) (*domain.PaymentTransaction, error)
	HandleCallback(ctx context.Context, params CallbackParams) error
	GetByID(ctx context.Context, id int64, merchantID int64) (*domain.PaymentTransaction, error)
	GetByOrderID(ctx context.Context, orderID int64, merchantID int64) (*domain.PaymentTransaction, error)
	GetStaticQR(ctx context.Context, merchantID int64) (*domain.StaticQR, error)
	UpdateStaticQR(ctx context.Context, qr *domain.StaticQR) error
	Sync(ctx context.Context, merchantID, branchID int64, lastSync *time.Time) (*PaymentSyncResult, error)
}

// PaymentSyncResult is the incremental payment payload for offline sync.
type PaymentSyncResult struct {
	Payments  []domain.PaymentTransaction `json:"payments"`
	SyncToken string                      `json:"syncToken"`
}

// CreatePaymentParams comes from the public payment API.
type CreatePaymentParams struct {
	MerchantID int64
	OrderID    int64
	Method     string
	BuyerName  string
	BuyerEmail string
	BuyerPhone string
}

// CallbackParams is the gateway webhook payload.
type CallbackParams struct {
	ReferenceID string
	Status      string
	StatusCode  int
	Amount      string
	PaymentRef  string // trx_id from gateway
	Via         string // va / qris / ew / cod
	Channel     string
	Signature   string
	MerchantVA  string
}
