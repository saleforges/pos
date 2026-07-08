package memory

import (
	"context"
	"sync"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
)

var _ repository.VariantRepository = (*VariantRepository)(nil)

type VariantRepository struct {
	mu       sync.RWMutex
	variants map[string]*domain.Variant
}

func NewVariantRepository() *VariantRepository {
	return &VariantRepository{
		variants: make(map[string]*domain.Variant),
	}
}

func (r *VariantRepository) Create(_ context.Context, variant *domain.Variant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.variants[variant.ID] = variant
	return nil
}

func (r *VariantRepository) GetByID(_ context.Context, id string) (*domain.Variant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.variants[id]
	if !ok {
		return nil, domain.ErrVariantNotFound
	}
	return v, nil
}

func (r *VariantRepository) ListByProduct(_ context.Context, productID string) ([]domain.Variant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Variant
	for _, v := range r.variants {
		if v.ProductID == productID {
			result = append(result, *v)
		}
	}
	return result, nil
}

func (r *VariantRepository) Update(_ context.Context, variant *domain.Variant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.variants[variant.ID] = variant
	return nil
}

func (r *VariantRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.variants, id)
	return nil
}
