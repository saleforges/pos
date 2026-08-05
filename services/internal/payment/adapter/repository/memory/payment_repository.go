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
	mu        sync.RWMutex
	payments  map[int64]*domain.PaymentTransaction
	staticQRs map[int64]*domain.StaticQR
	seq       int64
}

func NewPaymentRepository() *PaymentRepository {
	return &PaymentRepository{
		payments:  make(map[int64]*domain.PaymentTransaction),
		staticQRs: make(map[int64]*domain.StaticQR),
	}
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
	existing.PaymentRef = p.PaymentRef
	existing.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *PaymentRepository) GetByPaymentRef(_ context.Context, paymentRef string) (*domain.PaymentTransaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.payments {
		if p.PaymentRef == paymentRef {
			return p, nil
		}
	}
	return nil, domain.ErrPaymentNotFound
}

func (r *PaymentRepository) MarkPaid(_ context.Context, id int64, paymentRef string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.payments[id]
	if !ok {
		return domain.ErrPaymentNotFound
	}
	p.Status = domain.PaymentStatusPaid
	if paymentRef != "" {
		p.PaymentRef = paymentRef
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

func (r *PaymentRepository) GetStaticQR(_ context.Context, merchantID int64) (*domain.StaticQR, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, qr := range r.staticQRs {
		if qr.MerchantID == merchantID {
			return qr, nil
		}
	}
	return nil, domain.ErrPaymentNotFound
}

func (r *PaymentRepository) UpsertStaticQR(_ context.Context, qr *domain.StaticQR) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.staticQRs[qr.MerchantID] = qr
	return nil
}

func (r *PaymentRepository) ListChangedSince(_ context.Context, merchantID int64, since *time.Time) ([]domain.PaymentTransaction, error) {
	return r.syncByBranch(merchantID, 0, since)
}

func (r *PaymentRepository) SyncByBranch(_ context.Context, merchantID, branchID int64, since *time.Time) ([]domain.PaymentTransaction, error) {
	return r.syncByBranch(merchantID, branchID, since)
}

func (r *PaymentRepository) syncByBranch(merchantID, branchID int64, since *time.Time) ([]domain.PaymentTransaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.PaymentTransaction
	for _, p := range r.payments {
		if p.MerchantID != merchantID {
			continue
		}
		if branchID > 0 && p.BranchID != branchID {
			continue
		}
		if since != nil && !p.UpdatedAt.After(*since) {
			continue
		}
		result = append(result, *p)
	}
	return result, nil
}
