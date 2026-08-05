package repository

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/order/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Order, error)
	GetByClientOrderID(ctx context.Context, clientOrderID string, merchantID int64) (*domain.Order, error)
	List(ctx context.Context, merchantID int64, branchID *int64, status *domain.OrderStatus, paymentStatus *domain.PaymentStatus) ([]domain.Order, error)
	Update(ctx context.Context, order *domain.Order) error
	UpdateStatus(ctx context.Context, id int64, merchantID int64, status domain.OrderStatus) (*domain.Order, error)
	AddPayment(ctx context.Context, orderID int64, merchantID int64, payment *domain.PaymentRecord) error
	SalesReport(ctx context.Context, merchantID, branchID int64, from, to *time.Time) (*domain.SalesReport, error)
}
