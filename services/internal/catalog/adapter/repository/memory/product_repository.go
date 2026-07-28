package memory

import (
	"context"
	"sync"
	"time"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
)

var _ repository.ProductRepository = (*ProductRepository)(nil)

type ProductRepository struct {
	mu       sync.RWMutex
	products map[int64]*domain.Product
	seq      int64
}

func NewProductRepository() *ProductRepository {
	return &ProductRepository{
		products: make(map[int64]*domain.Product),
	}
}

func (r *ProductRepository) Create(_ context.Context, product *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	product.ID = r.seq
	r.products[product.ID] = product
	return nil
}

func (r *ProductRepository) GetByID(_ context.Context, id int64, merchantID int64) (*domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.products[id]
	if !ok || p.DeletedAt != nil || p.MerchantID != merchantID {
		return nil, domain.ErrProductNotFound
	}
	return p, nil
}

func (r *ProductRepository) List(_ context.Context, merchantID int64, search string, offset, limit int) ([]domain.Product, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.Product
	for _, p := range r.products {
		if p.MerchantID != merchantID || p.DeletedAt != nil {
			continue
		}
		if search != "" && !contains(p.Name, search) {
			continue
		}
		filtered = append(filtered, *p)
	}
	total := len(filtered)
	if offset >= total {
		return []domain.Product{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return filtered[offset:end], total, nil
}

func (r *ProductRepository) Update(_ context.Context, product *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.products[product.ID]
	if !ok || existing.MerchantID != product.MerchantID {
		return domain.ErrProductNotFound
	}
	r.products[product.ID] = product
	return nil
}

func (r *ProductRepository) Delete(_ context.Context, id int64, merchantID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p, ok := r.products[id]; ok && p.MerchantID == merchantID {
		now := time.Now().UTC()
		p.DeletedAt = &now
	}
	return nil
}

func (r *ProductRepository) Restore(_ context.Context, id int64, merchantID int64) (*domain.Product, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.products[id]
	if !ok || p.MerchantID != merchantID {
		return nil, domain.ErrProductNotFound
	}
	p.DeletedAt = nil
	return p, nil
}

func (r *ProductRepository) ListUpdatedAfter(_ context.Context, merchantID int64, after time.Time) ([]domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.Product
	for _, p := range r.products {
		if p.MerchantID != merchantID {
			continue
		}
		if p.UpdatedAt.After(after) || (p.DeletedAt != nil && p.DeletedAt.After(after)) {
			result = append(result, *p)
		}
	}
	if result == nil {
		return []domain.Product{}, nil
	}
	return result, nil
}

func contains(s, substr string) bool {
	return len(substr) == 0 || s != "" && containsSubstring(s, substr)
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
