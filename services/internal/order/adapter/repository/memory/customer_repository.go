package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/saleforge/pos/services/internal/order/domain"
	"github.com/saleforge/pos/services/internal/order/port/repository"
)

var _ repository.CustomerRepository = (*CustomerRepository)(nil)

type CustomerRepository struct {
	mu        sync.RWMutex
	customers map[int64]*domain.Customer
	seq       int64
}

func NewCustomerRepository() *CustomerRepository {
	return &CustomerRepository{
		customers: make(map[int64]*domain.Customer),
	}
}

func (r *CustomerRepository) Create(_ context.Context, c *domain.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.seq++
	c.ID = r.seq
	r.customers[c.ID] = c
	return nil
}

func (r *CustomerRepository) GetByID(_ context.Context, id int64, merchantID int64) (*domain.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.customers[id]
	if !ok || c.MerchantID != merchantID {
		return nil, domain.ErrCustomerNotFound
	}
	return c, nil
}

func (r *CustomerRepository) List(_ context.Context, merchantID int64, search string) ([]domain.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Customer
	for _, c := range r.customers {
		if c.MerchantID != merchantID {
			continue
		}
		if search != "" {
			needle := strings.ToLower(search)
			if !strings.Contains(strings.ToLower(c.Name), needle) && !strings.Contains(strings.ToLower(c.Phone), needle) {
				continue
			}
		}
		result = append(result, *c)
	}
	if result == nil {
		return []domain.Customer{}, nil
	}
	return result, nil
}

func (r *CustomerRepository) ListChangedSince(_ context.Context, merchantID int64, since *time.Time) ([]domain.Customer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Customer
	for _, c := range r.customers {
		if c.MerchantID != merchantID {
			continue
		}
		if since != nil && !c.UpdatedAt.After(*since) {
			continue
		}
		result = append(result, *c)
	}
	if result == nil {
		return []domain.Customer{}, nil
	}
	return result, nil
}

func (r *CustomerRepository) Update(_ context.Context, c *domain.Customer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.customers[c.ID]
	if !ok || existing.MerchantID != c.MerchantID {
		return domain.ErrCustomerNotFound
	}
	r.customers[c.ID] = c
	return nil
}

func (r *CustomerRepository) Delete(_ context.Context, id int64, merchantID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	c, ok := r.customers[id]
	if !ok || c.MerchantID != merchantID {
		return domain.ErrCustomerNotFound
	}
	delete(r.customers, id)
	return nil
}

func (r *CustomerRepository) UpsertPrices(_ context.Context, merchantID, customerID int64, prices []domain.CustomerPrice) error {
	return nil
}

func (r *CustomerRepository) ListPrices(_ context.Context, merchantID, customerID int64) ([]domain.CustomerPrice, error) {
	return []domain.CustomerPrice{}, nil
}

func (r *CustomerRepository) ListAllPrices(_ context.Context, merchantID int64) ([]domain.CustomerPrice, error) {
	return []domain.CustomerPrice{}, nil
}

func (r *CustomerRepository) GetPriceMap(_ context.Context, merchantID, customerID int64, productItemIDs []int64) (map[int64]float64, error) {
	return map[int64]float64{}, nil
}
