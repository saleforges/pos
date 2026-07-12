package memory

import (
	"context"
	"sync"

	"github.com/saleforge/pos/services/internal/merchant/domain"
)

type StaffRepository struct {
	mu    sync.RWMutex
	staff map[int64]*domain.StaffMember
	seq   int64
}

func NewStaffRepository() *StaffRepository {
	return &StaffRepository{staff: make(map[int64]*domain.StaffMember)}
}

func (r *StaffRepository) Create(_ context.Context, staff *domain.StaffMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	staff.ID = r.seq
	r.staff[staff.ID] = staff
	return nil
}

func (r *StaffRepository) GetByID(_ context.Context, id int64) (*domain.StaffMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.staff[id]
	if !ok {
		return nil, domain.ErrStaffNotFound
	}
	return s, nil
}

func (r *StaffRepository) ListByBranch(_ context.Context, branchID int64) ([]domain.StaffMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.StaffMember
	for _, s := range r.staff {
		if s.BranchID == branchID {
			result = append(result, *s)
		}
	}
	if result == nil {
		return []domain.StaffMember{}, nil
	}
	return result, nil
}

func (r *StaffRepository) ListByMerchant(_ context.Context, merchantID int64) ([]domain.StaffMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.StaffMember
	for _, s := range r.staff {
		if s.MerchantID == merchantID {
			result = append(result, *s)
		}
	}
	if result == nil {
		return []domain.StaffMember{}, nil
	}
	return result, nil
}

func (r *StaffRepository) GetByUserAndBranch(_ context.Context, userID, branchID int64) (*domain.StaffMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, s := range r.staff {
		if s.UserID == userID && s.BranchID == branchID {
			return s, nil
		}
	}
	return nil, domain.ErrStaffNotFound
}

func (r *StaffRepository) ListByUserAndMerchant(_ context.Context, userID, merchantID int64) ([]domain.StaffMember, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.StaffMember
	for _, s := range r.staff {
		if s.UserID == userID && s.MerchantID == merchantID {
			result = append(result, *s)
		}
	}
	if result == nil {
		return []domain.StaffMember{}, nil
	}
	return result, nil
}

func (r *StaffRepository) SetDefaultBranch(_ context.Context, userID, branchID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.staff {
		if s.UserID == userID {
			s.IsDefault = s.BranchID == branchID
		}
	}
	return nil
}

func (r *StaffRepository) Update(_ context.Context, staff *domain.StaffMember) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.staff[staff.ID] = staff
	return nil
}

func (r *StaffRepository) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.staff, id)
	return nil
}
