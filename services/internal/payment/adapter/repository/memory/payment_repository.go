package memory

import (
	"context"
	"sync"
	"time"

	"github.com/saleforge/pos/services/internal/payment/domain"
	"github.com/saleforge/pos/services/internal/payment/port/repository"
)

var _ repository.PaymentRepository = (*PaymentRepository)(nil)

type PaymentRepository struct {
	mu       sync.RWMutex
	payments map[int64]*domain.PaymentTransaction
	seq      int64
}

func NewPaymentRepository() *PaymentRepository {
	return &PaymentRepository{payments: make(map[int64]*domain.PaymentTransaction)}
}

func (r *PaymentRepository) Create(_ context.Context, p *domain.PaymentTransaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.seq++
	p.ID = r.seq
	r.payments[p.ID] = p
	return nil
}

func (r *PaymentRepository) GetByID(_ context.Context, id int64, merchantID int64) (*domain.PaymentTransaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.payments[id]
	if !ok || p.MerchantID != merchantID {
		return nil, domain.ErrPaymentNotFound
	}
	return p, nil
}

func (r *PaymentRepository) GetByOrderID(_ context.Context, orderID int64) (*domain.PaymentTransaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var latest *domain.PaymentTransaction
	for _, p := range r.payments {
		if p.OrderID == orderID {
			if latest == nil || p.ID > latest.ID {
				latest = p
			}
		}
	}
	if latest == nil {
		return nil, domain.ErrPaymentNotFound
	}
	return latest, nil
}

func (r *PaymentRepository) UpdatePaymentURL(_ context.Context, id int64, paymentURL, sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.payments[id]
	if !ok {
		return domain.ErrPaymentNotFound
	}
	p.PaymentURL = paymentURL
	p.SessionID = sessionID
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *PaymentRepository) UpdateDetails(_ context.Context, id int64, p *domain.PaymentTransaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.payments[id]
	if !ok {
		return domain.ErrPaymentNotFound
	}
	existing.PaymentURL = p.PaymentURL
	existing.PaymentNo = p.PaymentNo
	existing.QrString = p.QrString
	existing.QrImage = p.QrImage
	existing.ExpiredAt = p.ExpiredAt
	existing.SessionID = p.SessionID
	existing.TransactionID = p.TransactionID
	existing.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *PaymentRepository) GetByGatewayRef(_ context.Context, gatewayRef string) (*domain.PaymentTransaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.payments {
		if p.GatewayRef == gatewayRef {
			return p, nil
		}
	}
	return nil, domain.ErrPaymentNotFound
}

func (r *PaymentRepository) MarkPaid(_ context.Context, id int64, gatewayRef string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.payments[id]
	if !ok {
		return domain.ErrPaymentNotFound
	}
	p.Status = domain.PaymentStatusPaid
	if gatewayRef != "" {
		p.GatewayRef = gatewayRef
	}
	p.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *PaymentRepository) MarkExpired(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.payments[id]
	if !ok {
		return domain.ErrPaymentNotFound
	}
	p.Status = domain.PaymentStatusExpired
	p.UpdatedAt = time.Now().UTC()
	return nil
}
