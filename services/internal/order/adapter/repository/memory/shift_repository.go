package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/saleforge/pos/services/internal/order/domain"
	"github.com/saleforge/pos/services/internal/order/port/repository"
)

var _ repository.ShiftRepository = (*ShiftRepository)(nil)

type ShiftRepository struct {
	mu        sync.RWMutex
	shifts    map[int64]*domain.Shift
	seq       int64
	orderRepo *OrderRepository
}

func NewShiftRepository(orderRepo *OrderRepository) *ShiftRepository {
	return &ShiftRepository{
		shifts:    make(map[int64]*domain.Shift),
		orderRepo: orderRepo,
	}
}

func (r *ShiftRepository) Create(_ context.Context, s *domain.Shift) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	s.ID = r.seq
	r.shifts[s.ID] = s
	return nil
}

func (r *ShiftRepository) GetByID(_ context.Context, id int64, merchantID int64) (*domain.Shift, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.shifts[id]
	if !ok || s.MerchantID != merchantID {
		return nil, domain.ErrShiftNotFound
	}
	return s, nil
}

func (r *ShiftRepository) GetOpenByBranch(_ context.Context, merchantID, branchID int64) (*domain.Shift, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.shifts {
		if s.MerchantID == merchantID && s.BranchID == branchID && s.Status == domain.ShiftStatusOpen {
			return s, nil
		}
	}
	return nil, domain.ErrShiftNotFound
}

func (r *ShiftRepository) Update(_ context.Context, s *domain.Shift) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.shifts[s.ID]
	if !ok || existing.MerchantID != s.MerchantID {
		return domain.ErrShiftNotFound
	}
	r.shifts[s.ID] = s
	return nil
}

func (r *ShiftRepository) List(_ context.Context, merchantID, branchID int64) ([]domain.Shift, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Shift
	for _, s := range r.shifts {
		if s.MerchantID != merchantID {
			continue
		}
		if branchID > 0 && s.BranchID != branchID {
			continue
		}
		result = append(result, *s)
	}
	if result == nil {
		result = []domain.Shift{}
	}
	return result, nil
}

func (r *ShiftRepository) SumCashPayments(_ context.Context, merchantID, branchID int64, from, to time.Time) (float64, error) {
	if r.orderRepo == nil {
		return 0, nil
	}
	r.orderRepo.mu.RLock()
	defer r.orderRepo.mu.RUnlock()
	var total float64
	for _, o := range r.orderRepo.orders {
		if o.MerchantID != merchantID || o.BranchID != branchID {
			continue
		}
		for _, p := range o.Payments {
			method := strings.ToLower(string(p.Method))
			if (method == "cash" || method == "tunai") && !p.PaidAt.Before(from) && p.PaidAt.Before(to) {
				total += p.Amount
			}
		}
	}
	return total, nil
}
