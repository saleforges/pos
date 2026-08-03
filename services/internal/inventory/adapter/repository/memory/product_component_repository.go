package memory

import (
	"context"
	"sync"

	"github.com/saleforge/pos/services/internal/inventory/domain"
	"github.com/saleforge/pos/services/internal/inventory/port/repository"
)

var _ repository.ProductComponentRepository = (*ProductComponentRepository)(nil)

type ProductComponentRepository struct {
	mu         sync.RWMutex
	components map[int64]*domain.ProductComponent
	itemSeq    int64
	seq        int64
}

func NewProductComponentRepository() *ProductComponentRepository {
	return &ProductComponentRepository{
		components: make(map[int64]*domain.ProductComponent),
	}
}

func (r *ProductComponentRepository) Create(_ context.Context, component *domain.ProductComponent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check uniqueness: one ProductComponent per ProductItem
	for _, existing := range r.components {
		if existing.ProductItemID == component.ProductItemID && existing.MerchantID == component.MerchantID {
			return domain.ErrComponentAlreadyExists
		}
	}

	r.seq++
	component.ID = r.seq

	// Assign item IDs
	for i := range component.Items {
		r.itemSeq++
		component.Items[i].ID = r.itemSeq
		component.Items[i].ProductComponentID = component.ID
	}

	r.components[component.ID] = component
	return nil
}

func (r *ProductComponentRepository) GetByProductItem(_ context.Context, productItemID int64, merchantID int64) (*domain.ProductComponent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, c := range r.components {
		if c.ProductItemID == productItemID && c.MerchantID == merchantID {
			return c, nil
		}
	}
	return nil, domain.ErrProductComponentNotFound
}

func (r *ProductComponentRepository) List(_ context.Context, merchantID int64) ([]domain.ProductComponent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.ProductComponent
	for _, c := range r.components {
		if c.MerchantID == merchantID {
			result = append(result, *c)
		}
	}
	if result == nil {
		return []domain.ProductComponent{}, nil
	}
	return result, nil
}

func (r *ProductComponentRepository) Update(_ context.Context, component *domain.ProductComponent) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.components[component.ID]
	if !ok || existing.MerchantID != component.MerchantID {
		return domain.ErrProductComponentNotFound
	}

	r.components[component.ID] = component
	return nil
}

func (r *ProductComponentRepository) Delete(_ context.Context, id int64, merchantID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.components[id]
	if !ok || c.MerchantID != merchantID {
		return domain.ErrProductComponentNotFound
	}
	delete(r.components, id)
	return nil
}
