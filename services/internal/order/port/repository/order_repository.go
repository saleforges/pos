package repository

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/order/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Order, error)
	List(ctx context.Context, merchantID int64, branchID *int64, status *domain.OrderStatus, paymentStatus *domain.PaymentStatus) ([]domain.Order, error)
	UpdateStatus(ctx context.Context, id int64, merchantID int64, status domain.OrderStatus) (*domain.Order, error)
	UpdateDueDate(ctx context.Context, id int64, merchantID int64, dueDate *time.Time) (*domain.Order, error)
	AddPayment(ctx context.Context, orderID int64, merchantID int64, payment *domain.PaymentRecord) error
}
