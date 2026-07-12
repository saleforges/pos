package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
)

var _ repository.ProductRepository = (*ProductRepository)(nil)

type ProductRepository struct {
	mu   sync.RWMutex
	data map[int64]*domain.Product
	seq  int64
}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{
		data: make(map[int64]*domain.Product),
	}
}

func (r *ProductRepository) Create(ctx context.Context, product *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	product.ID = r.seq
	cp := *product
	r.data[product.ID] = &cp
	return nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.data[id]
	if !ok {
		return nil, fmt.Errorf("product not found")
	}
	cp := *p
	return &cp, nil
}

func (r *ProductRepository) GetBySKU(ctx context.Context, sku string, merchantID int64) (*domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.data {
		if p.SKU == sku && p.MerchantID == merchantID {
			cp := *p
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("product not found")
}

func (r *ProductRepository) List(ctx context.Context, merchantID int64, search string, offset, limit int) ([]domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.Product
	for _, p := range r.data {
		if p.MerchantID != merchantID {
			continue
		}
		if search != "" {
			search = strings.ToLower(search)
			if !strings.Contains(strings.ToLower(p.Name), search) &&
				!strings.Contains(strings.ToLower(p.SKU), search) &&
				!strings.Contains(strings.ToLower(p.Barcode), search) &&
				!strings.Contains(strings.ToLower(p.Description), search) {
				continue
			}
		}
		filtered = append(filtered, *p)
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

func (r *ProductRepository) ListByCategory(ctx context.Context, categoryID int64, offset, limit int) ([]domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.Product
	for _, p := range r.data {
		if p.CategoryID == categoryID {
			filtered = append(filtered, *p)
		}
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

func (r *ProductRepository) Count(ctx context.Context, merchantID int64, search string) (int, error) {
	list, _ := r.List(ctx, merchantID, search, 0, 1<<30)
	return len(list), nil
}

func (r *ProductRepository) Update(ctx context.Context, product *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.data[product.ID]
	if !ok {
		return fmt.Errorf("product not found")
	}
	cp := *product
	r.data[product.ID] = &cp
	return nil
}

func (r *ProductRepository) Delete(ctx context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, id)
	return nil
}
