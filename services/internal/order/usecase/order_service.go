package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/order/domain"
	"github.com/saleforge/pos/services/internal/order/port/repository"
)

type orderUsecase struct {
	orderRepo    repository.OrderRepository
	customerRepo repository.CustomerRepository
}

func NewOrderUsecase(orderRepo repository.OrderRepository, customerRepo repository.CustomerRepository) OrderUsecase {
	return &orderUsecase{orderRepo: orderRepo, customerRepo: customerRepo}
}

func (uc *orderUsecase) Create(ctx context.Context, params CreateOrderParams) (*domain.Order, error) {
	// Validate customer belongs to this merchant if provided
	if params.CustomerID != nil {
		if _, err := uc.customerRepo.GetByID(ctx, *params.CustomerID, params.MerchantID); err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC()
	items := make([]domain.OrderItem, len(params.Items))
	var subtotal float64
	for i, p := range params.Items {
		items[i] = domain.OrderItem{
			ProductItemID: p.ProductItemID,
			ItemName:      p.ItemName,
			UnitPrice:     p.UnitPrice,
			Quantity:      p.Quantity,
			LineTotal:     p.UnitPrice * p.Quantity,
		}
		subtotal += items[i].LineTotal
	}

	order := &domain.Order{
		MerchantID: params.MerchantID,
		BranchID:   params.BranchID,
		CreatedBy:  params.CreatedBy,
		CustomerID: params.CustomerID,
		Status:     domain.OrderStatusCompleted,
		Subtotal:   subtotal,
		Discount:   0,
		Tax:        0,
		Total:      subtotal,
		PaidAmount: 0,
		DueDate:    params.DueDate,
		Note:       params.Note,
		CreatedAt:  now,
		UpdatedAt:  now,
		Items:      items,
	}
	// Default due date: credit sales (customer attached) without an explicit
	// due date get H+7 from today.
	if order.CustomerID != nil && order.DueDate == nil {
		d := now.AddDate(0, 0, defaultDueDays)
		order.DueDate = &d
	}
	order.PaymentStatus = order.ComputePaymentStatus()

	if err := order.Validate(); err != nil {
		return nil, err
	}

	if err := uc.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

func (uc *orderUsecase) GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Order, error) {
	return uc.orderRepo.GetByID(ctx, id, merchantID)
}

func (uc *orderUsecase) List(ctx context.Context, merchantID int64, branchID *int64, status *domain.OrderStatus, paymentStatus *domain.PaymentStatus) ([]domain.Order, error) {
	return uc.orderRepo.List(ctx, merchantID, branchID, status, paymentStatus)
}

func (uc *orderUsecase) Cancel(ctx context.Context, id int64, merchantID int64) (*domain.Order, error) {
	order, err := uc.orderRepo.GetByID(ctx, id, merchantID)
	if err != nil {
		return nil, err
	}
	if order.Status != domain.OrderStatusCompleted {
		return nil, domain.ErrInvalidTransition
	}
	return uc.orderRepo.UpdateStatus(ctx, id, merchantID, domain.OrderStatusCancelled)
}

func (uc *orderUsecase) UpdateDueDate(ctx context.Context, id int64, merchantID int64, dueDate *time.Time) (*domain.Order, error) {
	order, err := uc.orderRepo.GetByID(ctx, id, merchantID)
	if err != nil {
		return nil, err
	}
	if order.Status != domain.OrderStatusCompleted {
		return nil, domain.ErrInvalidTransition
	}
	return uc.orderRepo.UpdateDueDate(ctx, id, merchantID, dueDate)
}

func (uc *orderUsecase) AddPayment(ctx context.Context, params AddPaymentParams) (*domain.Order, error) {
	if params.Amount <= 0 {
		return nil, domain.ErrInvalidPayment
	}
	switch params.Method {
	case domain.PaymentMethodCash, domain.PaymentMethodTransfer, domain.PaymentMethodQRIS:
	default:
		return nil, domain.ErrInvalidPayment
	}

	order, err := uc.orderRepo.GetByID(ctx, params.OrderID, params.MerchantID)
	if err != nil {
		return nil, err
	}
	if order.Status != domain.OrderStatusCompleted {
		return nil, domain.ErrInvalidTransition
	}

	remaining := order.Total - order.PaidAmount
	if params.Amount > remaining {
		return nil, domain.ErrPaymentExceedsTotal
	}

	payment := &domain.PaymentRecord{
		OrderID:   order.ID,
		Amount:    params.Amount,
		Method:    params.Method,
		CreatedBy: params.CreatedBy,
		PaidAt:    params.PaidAt,
		CreatedAt: time.Now().UTC(),
	}
	if err := uc.orderRepo.AddPayment(ctx, order.ID, params.MerchantID, payment); err != nil {
		return nil, err
	}
	return uc.orderRepo.GetByID(ctx, order.ID, params.MerchantID)
}

var _ OrderUsecase = (*orderUsecase)(nil)

const defaultDueDays = 7
