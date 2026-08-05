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
	inventory    repository.InventoryClient
}

func NewOrderUsecase(orderRepo repository.OrderRepository, customerRepo repository.CustomerRepository, inventory repository.InventoryClient) OrderUsecase {
	return &orderUsecase{orderRepo: orderRepo, customerRepo: customerRepo, inventory: inventory}
}

func (uc *orderUsecase) Create(ctx context.Context, params CreateOrderParams) (*domain.Order, error) {
	// Idempotency: a client-generated UUID identifies the order across
	// retries (offline-first mobile). If the same client order id was
	// already created, return the existing order instead of duplicating.
	if params.ClientOrderID != "" {
		if existing, err := uc.orderRepo.GetByClientOrderID(ctx, params.ClientOrderID, params.MerchantID); err == nil && existing != nil {
			return existing, nil
		}
	}

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
		MerchantID:    params.MerchantID,
		BranchID:      params.BranchID,
		CreatedBy:     params.CreatedBy,
		CustomerID:    params.CustomerID,
		ClientOrderID: params.ClientOrderID,
		Status:        domain.OrderStatusCompleted,
		Subtotal:      subtotal,
		Discount:      0,
		Tax:           0,
		Total:         subtotal,
		PaidAmount:    0,
		DueDate:       params.DueDate,
		Note:          params.Note,
		CreatedAt:     now,
		UpdatedAt:     now,
		Items:         items,
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
		// Concurrent retry with the same client order id: the unique index
		// rejected the duplicate insert — return the winner instead.
		if params.ClientOrderID != "" {
			if existing, getErr := uc.orderRepo.GetByClientOrderID(ctx, params.ClientOrderID, params.MerchantID); getErr == nil && existing != nil {
				return existing, nil
			}
		}
		return nil, err
	}

	// Completed orders deduct stock from inventory. If any line has
	// insufficient stock the order is cancelled and the error returned.
	if err := uc.deductOrder(ctx, order); err != nil {
		_, _ = uc.orderRepo.UpdateStatus(ctx, order.ID, order.MerchantID, domain.OrderStatusCancelled)
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
	// Restore the stock that was deducted when the order completed.
	if err := uc.restoreOrder(ctx, order); err != nil {
		return nil, err
	}
	return uc.orderRepo.UpdateStatus(ctx, id, merchantID, domain.OrderStatusCancelled)
}

func (uc *orderUsecase) deductOrder(ctx context.Context, order *domain.Order) error {
	items := make([]repository.StockAdjustmentItem, len(order.Items))
	for i, it := range order.Items {
		items[i] = repository.StockAdjustmentItem{ProductItemID: it.ProductItemID, Quantity: int64(it.Quantity)}
	}
	return uc.inventory.DeductStock(ctx, order.MerchantID, order.BranchID, "order", order.ID, items)
}

func (uc *orderUsecase) restoreOrder(ctx context.Context, order *domain.Order) error {
	items := make([]repository.StockAdjustmentItem, len(order.Items))
	for i, it := range order.Items {
		items[i] = repository.StockAdjustmentItem{ProductItemID: it.ProductItemID, Quantity: int64(it.Quantity)}
	}
	return uc.inventory.RestoreStock(ctx, order.MerchantID, order.BranchID, "order", order.ID, items)
}

// NotifyPaid is called by the payment service when the gateway confirms a
// successful payment. Idempotent: already-settled orders are a no-op.
func (uc *orderUsecase) NotifyPaid(ctx context.Context, params NotifyPaidParams) error {
	order, err := uc.orderRepo.GetByID(ctx, params.OrderID, params.MerchantID)
	if err != nil {
		return err
	}
	if order.PaidAmount >= order.Total {
		return nil
	}

	amount := params.Amount
	remaining := order.Total - order.PaidAmount
	if amount > remaining {
		amount = remaining
	}

	return uc.orderRepo.AddPayment(ctx, order.ID, order.MerchantID, &domain.PaymentRecord{
		Amount:    amount,
		Method:    domain.PaymentMethod(params.Method),
		CreatedBy: 0, // system (gateway callback via payment service)
		PaidAt:    time.Now().UTC(),
		CreatedAt: time.Now().UTC(),
	})
}

func (uc *orderUsecase) Update(ctx context.Context, params UpdateOrderParams) (*domain.Order, error) {
	order, err := uc.orderRepo.GetByID(ctx, params.ID, params.MerchantID)
	if err != nil {
		return nil, err
	}
	if order.Status != domain.OrderStatusCompleted {
		return nil, domain.ErrInvalidTransition
	}
	if params.DueDate != nil {
		order.DueDate = params.DueDate
	}
	if params.Note != nil {
		order.Note = *params.Note
	}
	order.UpdatedAt = time.Now().UTC()
	if err := uc.orderRepo.Update(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
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

func (uc *orderUsecase) SalesReport(ctx context.Context, merchantID, branchID int64, from, to *time.Time) (*domain.SalesReport, error) {
	return uc.orderRepo.SalesReport(ctx, merchantID, branchID, from, to)
}

var _ OrderUsecase = (*orderUsecase)(nil)

const defaultDueDays = 7
