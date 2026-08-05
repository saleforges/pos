package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type mockProductRepo struct {
	products map[int64]*domain.Product
	err      error
}

func (m *mockProductRepo) Create(_ context.Context, p *domain.Product) error {
	if m.err != nil {
		return m.err
	}
	if m.products == nil {
		m.products = make(map[int64]*domain.Product)
	}
	p.ID = int64(len(m.products) + 1)
	m.products[p.ID] = p
	return nil
}

func (m *mockProductRepo) GetByID(_ context.Context, id int64, merchantID int64) (*domain.Product, error) {
	if m.err != nil {
		return nil, m.err
	}
	p, ok := m.products[id]
	if !ok || p.MerchantID != merchantID {
		return nil, domain.ErrProductNotFound
	}
	return p, nil
}

func (m *mockProductRepo) List(_ context.Context, merchantID int64, search string, offset, limit int) ([]domain.Product, int, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	var filtered []domain.Product
	for _, p := range m.products {
		if p.MerchantID == merchantID {
			filtered = append(filtered, *p)
		}
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

func (m *mockProductRepo) Update(_ context.Context, p *domain.Product) error {
	if m.err != nil {
		return m.err
	}
	existing, ok := m.products[p.ID]
	if !ok || existing.MerchantID != p.MerchantID {
		return domain.ErrProductNotFound
	}
	m.products[p.ID] = p
	return nil
}

func (m *mockProductRepo) Delete(_ context.Context, id int64, merchantID int64) error {
	if m.err != nil {
		return m.err
	}
	p, ok := m.products[id]
	if !ok || p.MerchantID != merchantID {
		return domain.ErrProductNotFound
	}
	delete(m.products, id)
	return nil
}

func (m *mockProductRepo) Restore(_ context.Context, id int64, merchantID int64) (*domain.Product, error) {
	if m.err != nil {
		return nil, m.err
	}
	p, ok := m.products[id]
	if !ok || p.MerchantID != merchantID {
		return nil, domain.ErrProductNotFound
	}
	return p, nil
}

func (m *mockProductRepo) ListUpdatedAfter(_ context.Context, merchantID int64, _ time.Time) ([]domain.Product, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.Product
	for _, p := range m.products {
		if p.MerchantID == merchantID {
			result = append(result, *p)
		}
	}
	return result, nil
}

type mockCategoryRepo struct {
	categories map[int64]*domain.Category
	err        error
}

func (m *mockCategoryRepo) Create(_ context.Context, c *domain.Category) error {
	if m.err != nil {
		return m.err
	}
	if m.categories == nil {
		m.categories = make(map[int64]*domain.Category)
	}
	c.ID = int64(len(m.categories) + 1)
	m.categories[c.ID] = c
	return nil
}

func (m *mockCategoryRepo) GetByID(_ context.Context, id int64, merchantID int64) (*domain.Category, error) {
	if m.err != nil {
		return nil, m.err
	}
	c, ok := m.categories[id]
	if !ok || c.MerchantID != merchantID {
		return nil, domain.ErrCategoryNotFound
	}
	return c, nil
}

func (m *mockCategoryRepo) ListByMerchant(_ context.Context, merchantID int64) ([]domain.Category, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.Category
	for _, c := range m.categories {
		if c.MerchantID == merchantID {
			result = append(result, *c)
		}
	}
	return result, nil
}

func (m *mockCategoryRepo) Update(_ context.Context, c *domain.Category) error {
	if m.err != nil {
		return m.err
	}
	existing, ok := m.categories[c.ID]
	if !ok || existing.MerchantID != c.MerchantID {
		return domain.ErrCategoryNotFound
	}
	m.categories[c.ID] = c
	return nil
}

func (m *mockCategoryRepo) Delete(_ context.Context, id int64, merchantID int64) error {
	if m.err != nil {
		return m.err
	}
	c, ok := m.categories[id]
	if !ok || c.MerchantID != merchantID {
		return domain.ErrCategoryNotFound
	}
	delete(m.categories, id)
	return nil
}

func (m *mockCategoryRepo) Restore(_ context.Context, id int64, merchantID int64) (*domain.Category, error) {
	if m.err != nil {
		return nil, m.err
	}
	c, ok := m.categories[id]
	if !ok || c.MerchantID != merchantID {
		return nil, domain.ErrCategoryNotFound
	}
	return c, nil
}

func (m *mockCategoryRepo) ListUpdatedAfter(_ context.Context, merchantID int64, _ time.Time) ([]domain.Category, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.Category
	for _, c := range m.categories {
		if c.MerchantID == merchantID {
			result = append(result, *c)
		}
	}
	return result, nil
}

type mockProductItemRepo struct {
	items map[int64]*domain.ProductItem
	err   error
}

func (m *mockProductItemRepo) Create(_ context.Context, item *domain.ProductItem) error {
	if m.err != nil {
		return m.err
	}
	if m.items == nil {
		m.items = make(map[int64]*domain.ProductItem)
	}

	// Check SKU uniqueness within merchant scope
	if item.SKU != "" {
		for _, existing := range m.items {
			if existing.SKU == item.SKU && existing.MerchantID == item.MerchantID && existing.DeletedAt == nil {
				return domain.ErrSKUDuplicate
			}
		}
	}

	item.ID = int64(len(m.items) + 1)
	m.items[item.ID] = item
	return nil
}

func (m *mockProductItemRepo) GetByID(_ context.Context, id int64, merchantID int64) (*domain.ProductItem, error) {
	if m.err != nil {
		return nil, m.err
	}
	item, ok := m.items[id]
	if !ok || item.MerchantID != merchantID {
		return nil, domain.ErrProductItemNotFound
	}
	return item, nil
}

func (m *mockProductItemRepo) ListByProduct(_ context.Context, productID int64, merchantID int64) ([]domain.ProductItem, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.ProductItem
	for _, item := range m.items {
		if item.ProductID == productID && item.MerchantID == merchantID {
			result = append(result, *item)
		}
	}
	return result, nil
}

func (m *mockProductItemRepo) ListByMerchant(_ context.Context, merchantID int64) ([]domain.ProductItem, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.ProductItem
	for _, item := range m.items {
		if item.MerchantID == merchantID {
			result = append(result, *item)
		}
	}
	return result, nil
}

func (m *mockProductItemRepo) Update(_ context.Context, item *domain.ProductItem) error {
	if m.err != nil {
		return m.err
	}
	existing, ok := m.items[item.ID]
	if !ok || existing.MerchantID != item.MerchantID {
		return domain.ErrProductItemNotFound
	}

	// Check SKU uniqueness within merchant scope (skip if same item)
	if item.SKU != "" {
		for _, other := range m.items {
			if other.ID != item.ID && other.SKU == item.SKU && other.MerchantID == item.MerchantID && other.DeletedAt == nil {
				return domain.ErrSKUDuplicate
			}
		}
	}

	m.items[item.ID] = item
	return nil
}

func (m *mockProductItemRepo) Delete(_ context.Context, id int64, merchantID int64) error {
	if m.err != nil {
		return m.err
	}
	item, ok := m.items[id]
	if !ok || item.MerchantID != merchantID {
		return domain.ErrProductItemNotFound
	}
	delete(m.items, id)
	return nil
}

func (m *mockProductItemRepo) Restore(_ context.Context, id int64, merchantID int64) (*domain.ProductItem, error) {
	if m.err != nil {
		return nil, m.err
	}
	item, ok := m.items[id]
	if !ok || item.MerchantID != merchantID {
		return nil, domain.ErrProductItemNotFound
	}
	return item, nil
}

func (m *mockProductItemRepo) ListUpdatedAfter(_ context.Context, merchantID int64, _ time.Time) ([]domain.ProductItem, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.ProductItem
	for _, item := range m.items {
		if item.MerchantID == merchantID {
			result = append(result, *item)
		}
	}
	return result, nil
}

type mockUnitRepo struct {
	units map[int64]*domain.Unit
	err   error
}

func newMockUnitRepo() *mockUnitRepo {
	return &mockUnitRepo{
		units: map[int64]*domain.Unit{
			1: {ID: 1, Code: "PCS", Name: "Piece"},
			2: {ID: 2, Code: "KG", Name: "Kilogram"},
		},
	}
}

func (m *mockUnitRepo) GetAll(_ context.Context) ([]domain.Unit, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.Unit
	for _, u := range m.units {
		result = append(result, *u)
	}
	return result, nil
}

func (m *mockUnitRepo) GetByID(_ context.Context, id int64) (*domain.Unit, error) {
	if m.err != nil {
		return nil, m.err
	}
	u, ok := m.units[id]
	if !ok {
		return nil, domain.ErrUnitNotFound
	}
	return u, nil
}

func (m *mockUnitRepo) GetByCode(_ context.Context, code string) (*domain.Unit, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, u := range m.units {
		if u.Code == code {
			return u, nil
		}
	}
	return nil, domain.ErrUnitNotFound
}

func (m *mockProductItemRepo) SetBranchPrice(_ context.Context, _, _ int64, _ float64, _ string) error {
	return nil
}
func (m *mockProductItemRepo) DeleteBranchPrice(_ context.Context, _, _ int64) error { return nil }
