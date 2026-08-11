package memory

import (
	"context"
	"sort"
	"sync"
	"time"

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

func (r *OrderRepository) GetByClientOrderID(_ context.Context, clientOrderID string, merchantID int64) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, o := range r.orders {
		if o.ClientOrderID == clientOrderID && o.MerchantID == merchantID {
			return o, nil
		}
	}
	return nil, domain.ErrOrderNotFound
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

func (r *OrderRepository) Update(_ context.Context, order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.orders[order.ID]
	if !ok || existing.MerchantID != order.MerchantID {
		return domain.ErrOrderNotFound
	}
	r.orders[order.ID] = order
	return nil
}

func (r *OrderRepository) UpdateStatus(_ context.Context, id int64, merchantID int64, status domain.OrderStatus) (*domain.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	o, ok := r.orders[id]
	if !ok || o.MerchantID != merchantID {
		return nil, domain.ErrOrderNotFound
	}
	o.Status = status
	o.UpdatedAt = time.Now().UTC()
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

func (r *OrderRepository) SalesReport(_ context.Context, merchantID, branchID int64, from, to *time.Time) (*domain.SalesReport, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	report := domain.SalesReport{TopProducts: []domain.ProductSales{}, PaymentBreakdown: []domain.PaymentMethodTotal{}}
	productTotals := map[int64]*domain.ProductSales{}
	methodTotals := map[string]*domain.PaymentMethodTotal{}
	for _, o := range r.orders {
		if o.MerchantID != merchantID || o.Status != domain.OrderStatusCompleted {
			continue
		}
		if branchID > 0 && o.BranchID != branchID {
			continue
		}
		if from != nil && o.CreatedAt.Before(*from) {
			continue
		}
		if to != nil && o.CreatedAt.After(*to) {
			continue
		}
		report.TotalOrders++
		report.TotalRevenue += o.PaidAmount
		if o.PaymentStatus == domain.PaymentStatusPaid {
			report.PaidOrders++
		} else {
			report.DebtOrders++
			report.Outstanding += o.Total - o.PaidAmount
		}
		for _, it := range o.Items {
			p, ok := productTotals[it.ProductItemID]
			if !ok {
				p = &domain.ProductSales{ProductItemID: it.ProductItemID, Name: it.ItemName}
				productTotals[it.ProductItemID] = p
			}
			p.Quantity += it.Quantity
			p.Revenue += it.LineTotal
		}
		for _, pay := range o.Payments {
			m, ok := methodTotals[string(pay.Method)]
			if !ok {
				m = &domain.PaymentMethodTotal{Method: string(pay.Method)}
				methodTotals[string(pay.Method)] = m
			}
			m.Amount += pay.Amount
			m.Count++
		}
	}
	for _, p := range productTotals {
		report.TopProducts = append(report.TopProducts, *p)
	}
	sort.Slice(report.TopProducts, func(i, j int) bool { return report.TopProducts[i].Quantity > report.TopProducts[j].Quantity })
	if len(report.TopProducts) > 5 {
		report.TopProducts = report.TopProducts[:5]
	}
	for _, m := range methodTotals {
		report.PaymentBreakdown = append(report.PaymentBreakdown, *m)
	}
	sort.Slice(report.PaymentBreakdown, func(i, j int) bool { return report.PaymentBreakdown[i].Amount > report.PaymentBreakdown[j].Amount })
	return &report, nil
}

func (r *OrderRepository) ListChangedSince(_ context.Context, merchantID, branchID int64, since *time.Time) ([]domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Order
	for _, o := range r.orders {
		if o.MerchantID != merchantID {
			continue
		}
		if branchID > 0 && o.BranchID != branchID {
			continue
		}
		if since != nil && !o.UpdatedAt.After(*since) {
			continue
		}
		o.PaymentStatus = o.ComputePaymentStatus()
		result = append(result, *o)
	}
	return result, nil
}
