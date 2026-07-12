package memory

import (
	"context"
	"sync"

	"github.com/saleforge/pos/services/internal/merchant/domain"
)

type BranchRepository struct {
	mu       sync.RWMutex
	branches map[int64]*domain.Branch
	seq      int64
}

func NewBranchRepository() *BranchRepository {
	return &BranchRepository{branches: make(map[int64]*domain.Branch)}
}

func (r *BranchRepository) Create(_ context.Context, branch *domain.Branch) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	branch.ID = r.seq
	r.branches[branch.ID] = branch
	return nil
}

func (r *BranchRepository) GetByID(_ context.Context, id int64) (*domain.Branch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.branches[id]
	if !ok {
		return nil, domain.ErrBranchNotFound
	}
	return b, nil
}

func (r *BranchRepository) ListByMerchant(_ context.Context, merchantID int64) ([]domain.Branch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Branch
	for _, b := range r.branches {
		if b.MerchantID == merchantID {
			result = append(result, *b)
		}
	}
	if result == nil {
		return []domain.Branch{}, nil
	}
	return result, nil
}

func (r *BranchRepository) Update(_ context.Context, branch *domain.Branch) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.branches[branch.ID] = branch
	return nil
}

func (r *BranchRepository) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.branches, id)
	return nil
}
