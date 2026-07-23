package usecase

import (
	"context"

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

func (m *mockProductRepo) GetByID(_ context.Context, id int64) (*domain.Product, error) {
	if m.err != nil {
		return nil, m.err
	}
	p, ok := m.products[id]
	if !ok {
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
	m.products[p.ID] = p
	return nil
}

func (m *mockProductRepo) Delete(_ context.Context, id int64) error {
	if m.err != nil {
		return m.err
	}
	delete(m.products, id)
	return nil
}

func (m *mockProductRepo) Restore(_ context.Context, id int64) (*domain.Product, error) {
	if m.err != nil {
		return nil, m.err
	}
	p, ok := m.products[id]
	if !ok {
		return nil, domain.ErrProductNotFound
	}
	return p, nil
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

func (m *mockCategoryRepo) GetByID(_ context.Context, id int64) (*domain.Category, error) {
	if m.err != nil {
		return nil, m.err
	}
	c, ok := m.categories[id]
	if !ok {
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
	m.categories[c.ID] = c
	return nil
}

func (m *mockCategoryRepo) Delete(_ context.Context, id int64) error {
	if m.err != nil {
		return m.err
	}
	delete(m.categories, id)
	return nil
}

func (m *mockCategoryRepo) Restore(_ context.Context, id int64) (*domain.Category, error) {
	if m.err != nil {
		return nil, m.err
	}
	c, ok := m.categories[id]
	if !ok {
		return nil, domain.ErrCategoryNotFound
	}
	return c, nil
}

type mockSellableItemRepo struct {
	items map[int64]*domain.SellableItem
	err   error
}

func (m *mockSellableItemRepo) Create(_ context.Context, item *domain.SellableItem) error {
	if m.err != nil {
		return m.err
	}
	if m.items == nil {
		m.items = make(map[int64]*domain.SellableItem)
	}
	item.ID = int64(len(m.items) + 1)
	m.items[item.ID] = item
	return nil
}

func (m *mockSellableItemRepo) GetByID(_ context.Context, id int64) (*domain.SellableItem, error) {
	if m.err != nil {
		return nil, m.err
	}
	item, ok := m.items[id]
	if !ok {
		return nil, domain.ErrSellableItemNotFound
	}
	return item, nil
}

func (m *mockSellableItemRepo) ListByProduct(_ context.Context, productID int64) ([]domain.SellableItem, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.SellableItem
	for _, item := range m.items {
		if item.ProductID == productID {
			result = append(result, *item)
		}
	}
	return result, nil
}

func (m *mockSellableItemRepo) Update(_ context.Context, item *domain.SellableItem) error {
	if m.err != nil {
		return m.err
	}
	m.items[item.ID] = item
	return nil
}

func (m *mockSellableItemRepo) Delete(_ context.Context, id int64) error {
	if m.err != nil {
		return m.err
	}
	delete(m.items, id)
	return nil
}

func (m *mockSellableItemRepo) Restore(_ context.Context, id int64) (*domain.SellableItem, error) {
	if m.err != nil {
		return nil, m.err
	}
	item, ok := m.items[id]
	if !ok {
		return nil, domain.ErrSellableItemNotFound
	}
	return item, nil
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
