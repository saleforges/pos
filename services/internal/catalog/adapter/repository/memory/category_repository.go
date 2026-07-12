package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
)

var _ repository.CategoryRepository = (*CategoryRepository)(nil)

type CategoryRepository struct {
	mu   sync.RWMutex
	data map[int64]*domain.Category
	seq  int64
}

func NewCategoryRepository() *CategoryRepository {
	return &CategoryRepository{
		data: make(map[int64]*domain.Category),
	}
}

func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	category.ID = r.seq
	cp := *category
	r.data[category.ID] = &cp
	return nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (*domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cat, ok := r.data[id]
	if !ok {
		return nil, fmt.Errorf("category not found")
	}
	cp := *cat
	return &cp, nil
}

func (r *CategoryRepository) List(ctx context.Context, merchantID int64, search string, offset, limit int) ([]domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.Category
	for _, cat := range r.data {
		if cat.MerchantID != merchantID {
			continue
		}
		if search != "" {
			search = strings.ToLower(search)
			if !strings.Contains(strings.ToLower(cat.Name), search) &&
				!strings.Contains(strings.ToLower(cat.Slug), search) &&
				!strings.Contains(strings.ToLower(cat.Description), search) {
				continue
			}
		}
		filtered = append(filtered, *cat)
	}

	if offset >= len(filtered) {
		return nil, nil
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[offset:end], nil
}

func (r *CategoryRepository) Count(ctx context.Context, merchantID int64, search string) (int, error) {
	list, _ := r.List(ctx, merchantID, search, 0, 1<<30)
	return len(list), nil
	// ponytail: O(n) scan, fine for dev/memory
}

func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.data[category.ID]
	if !ok {
		return fmt.Errorf("category not found")
	}
	cp := *category
	r.data[category.ID] = &cp
	return nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}
