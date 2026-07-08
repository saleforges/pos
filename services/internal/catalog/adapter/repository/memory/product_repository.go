package memory

import (
	"context"
	"sync"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
)

var _ repository.ProductRepository = (*ProductRepository)(nil)

type ProductRepository struct {
	mu       sync.RWMutex
	products map[string]*domain.Product
}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{
		products: make(map[string]*domain.Product),
	}
}

func (r *ProductRepository) Create(_ context.Context, product *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.products[product.ID] = product
	return nil
}

func (r *ProductRepository) GetByID(_ context.Context, id string) (*domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.products[id]
	if !ok {
		return nil, domain.ErrProductNotFound
	}
	return p, nil
}

func (r *ProductRepository) GetBySKU(_ context.Context, sku string, merchantID string) (*domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, p := range r.products {
		if p.SKU == sku && p.MerchantID == merchantID {
			return p, nil
		}
	}
	return nil, domain.ErrProductNotFound
}

func (r *ProductRepository) List(_ context.Context, merchantID string, search string, offset, limit int) ([]domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Product
	for _, p := range r.products {
		if p.MerchantID == merchantID {
			if search != "" && !matchProduct(p, search) {
				continue
			}
			result = append(result, *p)
		}
	}
	start := min(offset, len(result))
	end := min(start+limit, len(result))
	return result[start:end], nil
}

func (r *ProductRepository) Count(_ context.Context, merchantID string, search string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var count int
	for _, p := range r.products {
		if p.MerchantID == merchantID {
			if search != "" && !matchProduct(p, search) {
				continue
			}
			count++
		}
	}
	return count, nil
}

func matchProduct(p *domain.Product, search string) bool {
	return contains(p.Name, search) || contains(p.SKU, search) || contains(p.Barcode, search) || contains(p.Description, search)
}

func (r *ProductRepository) ListByCategory(_ context.Context, categoryID string, offset, limit int) ([]domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Product
	for _, p := range r.products {
		if p.CategoryID == categoryID {
			result = append(result, *p)
		}
	}
	start := min(offset, len(result))
	end := min(start+limit, len(result))
	return result[start:end], nil
}

func (r *ProductRepository) Update(_ context.Context, product *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.products[product.ID] = product
	return nil
}

func (r *ProductRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.products, id)
	return nil
}
