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
	staff []domain.StaffInfo // userID -> staff info (with userID embedded)
	index map[string][]int  // userID -> indices in staff slice
}

func NewStaffRepository() *StaffRepository {
	return &StaffRepository{
		staff: make([]domain.StaffInfo, 0),
		index: make(map[string][]int),
	}
}

func (r *StaffRepository) ListByUserID(_ context.Context, userID string) ([]domain.StaffInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	indices, ok := r.index[userID]
	if !ok {
		return nil, nil
	}
	result := make([]domain.StaffInfo, len(indices))
	for i, idx := range indices {
		result[i] = r.staff[idx]
	}
	return result, nil
}

func (r *StaffRepository) Create(_ context.Context, userID, merchantID, merchantName, role string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	info := domain.StaffInfo{
		MerchantID:   merchantID,
		MerchantName: merchantName,
		Role:         role,
	}
	r.staff = append(r.staff, info)
	r.index[userID] = append(r.index[userID], len(r.staff)-1)
	return nil
}
