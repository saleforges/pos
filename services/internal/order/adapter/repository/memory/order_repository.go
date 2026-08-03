package memory

import (
	"context"
	"sync"

	"github.com/saleforge/pos/services/internal/order/domain"
	"github.com/saleforge/pos/services/internal/order/port/repository"
)

var _ repository.OrderRepository = (*OrderRepository)(nil)

type OrderRepository struct {
	mu         sync.RWMutex
	orders     map[int64]*domain.Order
	itemSeq    int64
	paymentSeq int64
	seq        int64
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		orders: make(map[int64]*domain.Order),
	}
}

func (r *OrderRepository) Create(_ context.Context, order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.seq++
	order.ID = r.seq
	for i := range order.Items {
		r.itemSeq++
		order.Items[i].ID = r.itemSeq
		order.Items[i].OrderID = order.ID
	}
	r.orders[order.ID] = order
	return nil
}

func (r *OrderRepository) GetByID(_ context.Context, id int64, merchantID int64) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	o, ok := r.orders[id]
	if !ok || o.MerchantID != merchantID {
		return nil, domain.ErrOrderNotFound
	}
	return o, nil
}

func (r *OrderRepository) List(_ context.Context, merchantID int64, branchID *int64, status *domain.OrderStatus, paymentStatus *domain.PaymentStatus) ([]domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Order
	for _, o := range r.orders {
		if o.MerchantID != merchantID {
			continue
		}
		if branchID != nil && o.BranchID != *branchID {
			continue
		}
		if status != nil && o.Status != *status {
			continue
		}
		if paymentStatus != nil && o.ComputePaymentStatus() != *paymentStatus {
			continue
		}
		result = append(result, *o)
	}
	if result == nil {
		return []domain.Order{}, nil
	}
	return result, nil
}

func (r *OrderRepository) UpdateStatus(_ context.Context, id int64, merchantID int64, status domain.OrderStatus) (*domain.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	o, ok := r.orders[id]
	if !ok || o.MerchantID != merchantID {
		return nil, domain.ErrOrderNotFound
	}
	o.Status = status
	return o, nil
}

func (r *OrderRepository) AddPayment(_ context.Context, orderID int64, merchantID int64, payment *domain.PaymentRecord) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	o, ok := r.orders[orderID]
	if !ok || o.MerchantID != merchantID {
		return domain.ErrOrderNotFound
	}

	r.paymentSeq++
	payment.ID = r.paymentSeq
	payment.OrderID = orderID
	o.Payments = append(o.Payments, *payment)
	o.PaidAmount += payment.Amount
	o.PaymentStatus = o.ComputePaymentStatus()
	return nil
}
