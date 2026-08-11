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
	Update(ctx context.Context, params UpdateOrderParams) (*domain.Order, error)
	Cancel(ctx context.Context, params CancelOrderParams) (*domain.Order, error)
	AddPayment(ctx context.Context, params AddPaymentParams) (*domain.Order, error)
	NotifyPaid(ctx context.Context, params NotifyPaidParams) error
	SalesReport(ctx context.Context, merchantID, branchID int64, from, to *time.Time) (*domain.SalesReport, error)
	Receipt(ctx context.Context, id int64, merchantID int64) (*domain.Receipt, error)
	SyncOrders(ctx context.Context, merchantID, branchID int64, lastSync *time.Time) (*OrderSyncResult, error)
}

type OrderSyncResult struct {
	Orders    []domain.Order `json:"orders"`
	SyncToken string         `json:"syncToken"`
}

type NotifyPaidParams struct {
	OrderID    int64
	MerchantID int64
	Amount     float64
	Method     string
}

type CreateOrderParams struct {
	MerchantID    int64
	BranchID      int64
	CreatedBy     int64
	CustomerID    *int64
	ClientOrderID string
	DueDate       *time.Time
	Note          string
	Discount      float64
	Items         []CreateOrderItemParams
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

type CancelOrderParams struct {
	OrderID     int64
	MerchantID  int64
	CancelledBy int64
}

type UpdateOrderParams struct {
	ID         int64
	MerchantID int64
	DueDate    *time.Time
	Note       *string
}
