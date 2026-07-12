package memory

import (
	"context"
	"fmt"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
	"sync"
)

var _ repository.VariantRepository = (*VariantRepository)(nil)

type VariantRepository struct {
	mu   sync.RWMutex
	data map[int64]*domain.Variant
	seq  int64
}

func NewVariantRepository() *VariantRepository {
	return &VariantRepository{
		data: make(map[int64]*domain.Variant),
	}
}

func (r *VariantRepository) Create(ctx context.Context, variant *domain.Variant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	variant.ID = r.seq
	cp := *variant
	r.data[variant.ID] = &cp
	return nil
}

func (r *VariantRepository) GetByID(ctx context.Context, id int64) (*domain.Variant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.data[id]
	if !ok {
		return nil, fmt.Errorf("variant not found")
	}
	cp := *v
	return &cp, nil
}

func (r *VariantRepository) ListByProduct(ctx context.Context, productID int64) ([]domain.Variant, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Variant
	for _, v := range r.data {
		if v.ProductID == productID {
			result = append(result, *v)
		}
	}
	return result, nil
}

func (r *VariantRepository) CountByProduct(ctx context.Context, productID int64) (int, error) {
	list, _ := r.ListByProduct(ctx, productID)
	return len(list), nil
}

func (r *VariantRepository) Update(ctx context.Context, variant *domain.Variant) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.data[variant.ID]
	if !ok {
		return fmt.Errorf("variant not found")
	}
	cp := *variant
	r.data[variant.ID] = &cp
	return nil
}

func (r *VariantRepository) Delete(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}
