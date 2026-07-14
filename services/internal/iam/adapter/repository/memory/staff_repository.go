package memory

import (
	"context"
	"sync"

	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port/repository"
)

var _ repository.StaffRepository = (*StaffRepository)(nil)

type StaffRepository struct {
	mu    sync.RWMutex
	staff []domain.UserRoleAssignment
	index map[int64][]int
}

func NewStaffRepository() *StaffRepository {
	return &StaffRepository{
		staff: make([]domain.UserRoleAssignment, 0),
		index: make(map[int64][]int),
	}
}

func (r *StaffRepository) ListByUserID(_ context.Context, userID int64) ([]domain.UserRoleAssignment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	indices, ok := r.index[userID]
	if !ok {
		return nil, nil
	}
	result := make([]domain.UserRoleAssignment, len(indices))
	for i, idx := range indices {
		result[i] = r.staff[idx]
	}
	return result, nil
}

func (r *StaffRepository) SetDefaultRole(_ context.Context, userID, roleID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, idx := range r.index[userID] {
		r.staff[idx].IsDefault = r.staff[idx].Role.ID == roleID
	}
	return nil
}

func (r *StaffRepository) Create(_ context.Context, userID int64, merchantID int64, merchantName, role string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	a := domain.UserRoleAssignment{
		MerchantID:   merchantID,
		MerchantName: merchantName,
		Role:         domain.Role{Name: role},
	}
	r.staff = append(r.staff, a)
	r.index[userID] = append(r.index[userID], len(r.staff)-1)
	return nil
}
