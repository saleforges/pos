package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/payment/domain"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *domain.PaymentTransaction) error
	GetByID(ctx context.Context, id int64, merchantID int64) (*domain.PaymentTransaction, error)
	GetByOrderID(ctx context.Context, orderID int64) (*domain.PaymentTransaction, error)
	GetByPaymentRef(ctx context.Context, paymentRef string) (*domain.PaymentTransaction, error)
	UpdatePaymentURL(ctx context.Context, id int64, paymentURL, sessionID string) error
	UpdateDetails(ctx context.Context, id int64, p *domain.PaymentTransaction) error
	MarkPaid(ctx context.Context, id int64, paymentRef string) error
	MarkExpired(ctx context.Context, id int64) error
	GetStaticQR(ctx context.Context, merchantID int64) (*domain.StaticQR, error)
	UpsertStaticQR(ctx context.Context, qr *domain.StaticQR) error
}
