package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/order/domain"
	"github.com/saleforge/pos/services/internal/order/port/repository"
)

type mockInventoryClient struct {
	deductErr  error
	restoreErr error
	deducted   []repository.StockAdjustmentItem
	restored   []repository.StockAdjustmentItem
}

func (m *mockInventoryClient) DeductStock(_ context.Context, _ int64, _ int64, _ string, _ int64, items []repository.StockAdjustmentItem) error {
	if m.deductErr != nil {
		return m.deductErr
	}
	m.deducted = append(m.deducted, items...)
	return nil
}

func (m *mockInventoryClient) RestoreStock(_ context.Context, _ int64, _ int64, _ string, _ int64, items []repository.StockAdjustmentItem) error {
	if m.restoreErr != nil {
		return m.restoreErr
	}
	m.restored = append(m.restored, items...)
	return nil
}

type mockOrderRepo struct {
	orders  map[int64]*domain.Order
	itemSeq int64
	paySeq  int64
	seq     int64
	err     error
}

func (m *mockOrderRepo) Create(_ context.Context, order *domain.Order) error {
	if m.err != nil {
		return m.err
	}
	if m.orders == nil {
		m.orders = make(map[int64]*domain.Order)
	}
	m.seq++
	order.ID = m.seq
	for i := range order.Items {
		m.itemSeq++
		order.Items[i].ID = m.itemSeq
		order.Items[i].OrderID = order.ID
	}
	m.orders[order.ID] = order
	return nil
}

func (m *mockOrderRepo) GetByID(_ context.Context, id int64, merchantID int64) (*domain.Order, error) {
	if m.err != nil {
		return nil, m.err
	}
	o, ok := m.orders[id]
	if !ok || o.MerchantID != merchantID {
		return nil, domain.ErrOrderNotFound
	}
	o.PaymentStatus = o.ComputePaymentStatus()
	return o, nil
}

func (m *mockOrderRepo) GetByClientOrderID(_ context.Context, clientOrderID string, merchantID int64) (*domain.Order, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, o := range m.orders {
		if o.ClientOrderID == clientOrderID && o.MerchantID == merchantID {
			o.PaymentStatus = o.ComputePaymentStatus()
			return o, nil
		}
	}
	return nil, domain.ErrOrderNotFound
}

func (m *mockOrderRepo) List(_ context.Context, merchantID int64, branchID *int64, status *domain.OrderStatus, paymentStatus *domain.PaymentStatus) ([]domain.Order, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.Order
	for _, o := range m.orders {
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

func (m *mockOrderRepo) Update(_ context.Context, order *domain.Order) error {
	if m.err != nil {
		return m.err
	}
	existing, ok := m.orders[order.ID]
	if !ok || existing.MerchantID != order.MerchantID {
		return domain.ErrOrderNotFound
	}
	m.orders[order.ID] = order
	return nil
}

func (m *mockOrderRepo) UpdateStatus(_ context.Context, id int64, merchantID int64, status domain.OrderStatus) (*domain.Order, error) {
	if m.err != nil {
		return nil, m.err
	}
	o, ok := m.orders[id]
	if !ok || o.MerchantID != merchantID {
		return nil, domain.ErrOrderNotFound
	}
	o.Status = status
	o.UpdatedAt = time.Now().UTC()
	return o, nil
}

func (m *mockOrderRepo) AddPayment(_ context.Context, orderID int64, merchantID int64, payment *domain.PaymentRecord) error {
	if m.err != nil {
		return m.err
	}
	o, ok := m.orders[orderID]
	if !ok || o.MerchantID != merchantID {
		return domain.ErrOrderNotFound
	}
	m.paySeq++
	payment.ID = m.paySeq
	payment.OrderID = orderID
	o.Payments = append(o.Payments, *payment)
	o.PaidAmount += payment.Amount
	return nil
}

type mockCustomerRepo struct {
	customers map[int64]*domain.Customer
	seq       int64
	err       error
}

func (m *mockCustomerRepo) Create(_ context.Context, c *domain.Customer) error {
	if m.err != nil {
		return m.err
	}
	if m.customers == nil {
		m.customers = make(map[int64]*domain.Customer)
	}
	m.seq++
	c.ID = m.seq
	m.customers[c.ID] = c
	return nil
}

func (m *mockCustomerRepo) GetByID(_ context.Context, id int64, merchantID int64) (*domain.Customer, error) {
	if m.err != nil {
		return nil, m.err
	}
	c, ok := m.customers[id]
	if !ok || c.MerchantID != merchantID {
		return nil, domain.ErrCustomerNotFound
	}
	return c, nil
}

func (m *mockCustomerRepo) List(_ context.Context, merchantID int64, _ string) ([]domain.Customer, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.Customer
	for _, c := range m.customers {
		if c.MerchantID == merchantID {
			result = append(result, *c)
		}
	}
	if result == nil {
		return []domain.Customer{}, nil
	}
	return result, nil
}

func (m *mockCustomerRepo) ListChangedSince(_ context.Context, merchantID int64, since *time.Time) ([]domain.Customer, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.Customer
	for _, c := range m.customers {
		if c.MerchantID != merchantID {
			continue
		}
		if since != nil && !c.UpdatedAt.After(*since) {
			continue
		}
		result = append(result, *c)
	}
	if result == nil {
		return []domain.Customer{}, nil
	}
	return result, nil
}

func (m *mockCustomerRepo) Update(_ context.Context, c *domain.Customer) error {
	if m.err != nil {
		return m.err
	}
	existing, ok := m.customers[c.ID]
	if !ok || existing.MerchantID != c.MerchantID {
		return domain.ErrCustomerNotFound
	}
	m.customers[c.ID] = c
	return nil
}

func (m *mockCustomerRepo) Delete(_ context.Context, id int64, merchantID int64) error {
	if m.err != nil {
		return m.err
	}
	c, ok := m.customers[id]
	if !ok || c.MerchantID != merchantID {
		return domain.ErrCustomerNotFound
	}
	delete(m.customers, id)
	return nil
}

func newMockCustomerRepo() *mockCustomerRepo {
	repo := &mockCustomerRepo{}
	repo.Create(context.Background(), &domain.Customer{MerchantID: 1, Name: "Pak Budi"})
	return repo
}

func (m *mockOrderRepo) SalesReport(_ context.Context, _ int64, _ int64, _ *time.Time, _ *time.Time) (*domain.SalesReport, error) {
	return &domain.SalesReport{}, nil
}

func (m *mockOrderRepo) ListChangedSince(_ context.Context, merchantID, branchID int64, since *time.Time) ([]domain.Order, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.Order
	for _, o := range m.orders {
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
