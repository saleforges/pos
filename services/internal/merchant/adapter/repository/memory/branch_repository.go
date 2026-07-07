package memory

import (
	"context"
	"sync"

	"github.com/saleforge/pos/services/internal/merchant/domain"
)

type BranchRepository struct {
	mu     sync.RWMutex
	branches map[string]*domain.Branch
}

func NewBranchRepository() *BranchRepository {
	return &BranchRepository{branches: make(map[string]*domain.Branch)}
}

func (r *BranchRepository) Create(_ context.Context, branch *domain.Branch) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.branches[branch.ID] = branch
	return nil
}

func (r *BranchRepository) GetByID(_ context.Context, id string) (*domain.Branch, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.branches[id]
	if !ok {
		return nil, domain.ErrBranchNotFound
	}
	return b, nil
}

func (r *BranchRepository) ListByMerchant(_ context.Context, merchantID string) ([]domain.Branch, error) {
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

func (r *BranchRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.branches, id)
	return nil
}
