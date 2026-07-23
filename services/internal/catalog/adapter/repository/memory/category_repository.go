package memory

import (
	"context"
	"sync"
	"time"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
)

var _ repository.CategoryRepository = (*CategoryRepository)(nil)

type CategoryRepository struct {
	mu         sync.RWMutex
	categories map[int64]*domain.Category
	seq        int64
}

func NewCategoryRepository() *CategoryRepository {
	return &CategoryRepository{
		categories: make(map[int64]*domain.Category),
	}
}

func (r *CategoryRepository) Create(_ context.Context, category *domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	category.ID = r.seq
	r.categories[category.ID] = category
	return nil
}

func (r *CategoryRepository) GetByID(_ context.Context, id int64, merchantID int64) (*domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.categories[id]
	if !ok || c.DeletedAt != nil || c.MerchantID != merchantID {
		return nil, domain.ErrCategoryNotFound
	}
	return c, nil
}

func (r *CategoryRepository) ListByMerchant(_ context.Context, merchantID int64) ([]domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Category
	for _, c := range r.categories {
		if c.MerchantID == merchantID && c.DeletedAt == nil {
			result = append(result, *c)
		}
	}
	if result == nil {
		return []domain.Category{}, nil
	}
	return result, nil
}

func (r *CategoryRepository) Update(_ context.Context, category *domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.categories[category.ID]
	if !ok || existing.MerchantID != category.MerchantID {
		return domain.ErrCategoryNotFound
	}
	r.categories[category.ID] = category
	return nil
}

func (r *CategoryRepository) Delete(_ context.Context, id int64, merchantID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if c, ok := r.categories[id]; ok && c.MerchantID == merchantID {
		now := time.Now().UTC()
		c.DeletedAt = &now
	}
	return nil
}

func (r *CategoryRepository) Restore(_ context.Context, id int64, merchantID int64) (*domain.Category, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.categories[id]
	if !ok || c.MerchantID != merchantID {
		return nil, domain.ErrCategoryNotFound
	}
	c.DeletedAt = nil
	return c, nil
}
