package memory

import (
	"context"
	"sync"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
)

var _ repository.CategoryRepository = (*CategoryRepository)(nil)

type CategoryRepository struct {
	mu         sync.RWMutex
	categories map[string]*domain.Category
}

func NewCategoryRepository() *CategoryRepository {
	return &CategoryRepository{
		categories: make(map[string]*domain.Category),
	}
}

func (r *CategoryRepository) Create(_ context.Context, category *domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.categories[category.ID] = category
	return nil
}

func (r *CategoryRepository) GetByID(_ context.Context, id string) (*domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cat, ok := r.categories[id]
	if !ok {
		return nil, domain.ErrCategoryNotFound
	}
	return cat, nil
}

func (r *CategoryRepository) List(_ context.Context, merchantID string, search string, offset, limit int) ([]domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Category
	for _, cat := range r.categories {
		if cat.MerchantID == merchantID {
			if search != "" && !matchCategory(cat, search) {
				continue
			}
			result = append(result, *cat)
		}
	}
	start := min(offset, len(result))
	end := min(start+limit, len(result))
	return result[start:end], nil
}

func (r *CategoryRepository) Count(_ context.Context, merchantID string, search string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var count int
	for _, cat := range r.categories {
		if cat.MerchantID == merchantID {
			if search != "" && !matchCategory(cat, search) {
				continue
			}
			count++
		}
	}
	return count, nil
}

func matchCategory(cat *domain.Category, search string) bool {
	return contains(cat.Name, search) || contains(cat.Slug, search) || contains(cat.Description, search)
}

func (r *CategoryRepository) Update(_ context.Context, category *domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.categories[category.ID] = category
	return nil
}

func (r *CategoryRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.categories, id)
	return nil
}

func contains(s, substr string) bool {
	return len(substr) > 0 && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
