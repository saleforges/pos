package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/payment/domain"
	"github.com/saleforge/pos/services/internal/payment/port/repository"
)

type mockPaymentRepo struct {
	payments map[int64]*domain.PaymentTransaction
	seq      int64
}

func newMockPaymentRepo() *mockPaymentRepo {
	return &mockPaymentRepo{payments: map[int64]*domain.PaymentTransaction{}}
}

func (m *mockPaymentRepo) Create(_ context.Context, p *domain.PaymentTransaction) error {
	m.seq++
	p.ID = m.seq
	m.payments[p.ID] = p
	return nil
}

func (m *mockPaymentRepo) GetByID(_ context.Context, id int64, merchantID int64) (*domain.PaymentTransaction, error) {
	p, ok := m.payments[id]
	if !ok || p.MerchantID != merchantID {
		return nil, domain.ErrPaymentNotFound
	}
	return p, nil
}

func (m *mockPaymentRepo) GetByOrderID(_ context.Context, orderID int64) (*domain.PaymentTransaction, error) {
	for _, p := range m.payments {
		if p.OrderID == orderID {
			return p, nil
		}
	}
	return nil, domain.ErrPaymentNotFound
}

func (m *mockPaymentRepo) GetByPaymentRef(_ context.Context, paymentRef string) (*domain.PaymentTransaction, error) {
	for _, p := range m.payments {
		if p.PaymentRef == paymentRef {
			return p, nil
		}
	}
	return nil, domain.ErrPaymentNotFound
}

func (m *mockPaymentRepo) UpdatePaymentURL(_ context.Context, id int64, url, session string) error {
	p, ok := m.payments[id]
	if !ok {
		return domain.ErrPaymentNotFound
	}
	p.PaymentURL = url
	p.SessionID = session
	return nil
}

func (m *mockPaymentRepo) UpdateDetails(_ context.Context, id int64, p *domain.PaymentTransaction) error {
	existing, ok := m.payments[id]
	if !ok {
		return domain.ErrPaymentNotFound
	}
	existing.PaymentURL = p.PaymentURL
	existing.PaymentNo = p.PaymentNo
	existing.QrString = p.QrString
	existing.QrImage = p.QrImage
	existing.ExpiredAt = p.ExpiredAt
	existing.SessionID = p.SessionID
	existing.PaymentRef = p.PaymentRef
	return nil
}

func (m *mockPaymentRepo) MarkPaid(_ context.Context, id int64, paymentRef string) error {
	p, ok := m.payments[id]
	if !ok {
		return domain.ErrPaymentNotFound
	}
	p.Status = domain.PaymentStatusPaid
	p.PaymentRef = paymentRef
	return nil
}

func (m *mockPaymentRepo) MarkExpired(_ context.Context, id int64) error {
	p, ok := m.payments[id]
	if !ok {
		return domain.ErrPaymentNotFound
	}
	p.Status = domain.PaymentStatusExpired
	return nil
}

type mockGateway struct {
	result *repository.PaymentResult
	err    error
	called bool
}

func (m *mockGateway) CreatePayment(_ context.Context, _ repository.CreatePaymentParams) (*repository.PaymentResult, error) {
	m.called = true
	if m.err != nil {
		return nil, m.err
	}
	if m.result != nil {
		return m.result, nil
	}
	return &repository.PaymentResult{SessionID: "ses-1", PaymentURL: "https://sandbox.ipaymu.com/pay/1"}, nil
}

type mockOrderClient struct {
	notified bool
	err      error
	order    *repository.OrderInfo
	orderErr error
}

func (m *mockOrderClient) GetOrder(_ context.Context, orderID, _ int64) (*repository.OrderInfo, error) {
	if m.orderErr != nil {
		return nil, m.orderErr
	}
	if m.order != nil {
		return m.order, nil
	}
	return &repository.OrderInfo{
		ID:         orderID,
		MerchantID: 1,
		Status:     "completed",
		Total:      30000,
		Items:      []repository.OrderItem{{ItemName: "Es Teh", Quantity: 2, UnitPrice: 15000}},
	}, nil
}

func (m *mockOrderClient) NotifyPaid(_ context.Context, _, _ int64, _ float64, _ string) error {
	m.notified = true
	return m.err
}
