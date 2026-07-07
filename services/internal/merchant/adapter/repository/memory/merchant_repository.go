package memory

import (
	"context"
	"sync"

	"github.com/saleforge/pos/services/internal/merchant/domain"
)

type MerchantRepository struct {
	mu        sync.RWMutex
	merchants map[string]*domain.Merchant
}

func NewMerchantRepository() *MerchantRepository {
	return &MerchantRepository{merchants: make(map[string]*domain.Merchant)}
}

func (r *MerchantRepository) Create(_ context.Context, merchant *domain.Merchant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.merchants[merchant.ID] = merchant
	return nil
}

func (r *MerchantRepository) GetByID(_ context.Context, id string) (*domain.Merchant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.merchants[id]
	if !ok {
		return nil, domain.ErrMerchantNotFound
	}
	return m, nil
}

func (r *MerchantRepository) List(_ context.Context, offset, limit int) ([]domain.Merchant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	all := make([]domain.Merchant, 0, len(r.merchants))
	for _, m := range r.merchants {
		all = append(all, *m)
	}
	if offset >= len(all) {
		return []domain.Merchant{}, nil
	}
	end := offset + limit
	if end > len(all) {
		end = len(all)
	}
	return all[offset:end], nil
}

func (r *MerchantRepository) Update(_ context.Context, merchant *domain.Merchant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.merchants[merchant.ID] = merchant
	return nil
}

func (r *MerchantRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.merchants, id)
	return nil
}
