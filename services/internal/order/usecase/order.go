package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/order/domain"
)

type OrderUsecase interface {
	Create(ctx context.Context, params CreateOrderParams) (*domain.Order, error)
	GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Order, error)
	List(ctx context.Context, merchantID int64, branchID *int64, status *domain.OrderStatus, paymentStatus *domain.PaymentStatus) ([]domain.Order, error)
	Cancel(ctx context.Context, id int64, merchantID int64) (*domain.Order, error)
	AddPayment(ctx context.Context, params AddPaymentParams) (*domain.Order, error)
}

type CreateOrderParams struct {
	MerchantID int64
	BranchID   int64
	CreatedBy  int64
	CustomerID *int64
	DueDate    *time.Time
	Note       string
	Items      []CreateOrderItemParams
}

type CreateOrderItemParams struct {
	ProductItemID int64
	ItemName      string
	UnitPrice     float64
	Quantity      float64
}

type AddPaymentParams struct {
	OrderID    int64
	MerchantID int64
	CreatedBy  int64
	Amount     float64
	Method     domain.PaymentMethod
	PaidAt     time.Time
}
