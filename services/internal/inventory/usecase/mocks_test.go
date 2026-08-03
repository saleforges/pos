package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/inventory/domain"
	"github.com/saleforge/pos/services/internal/inventory/port/repository"
)

// Mock repositories for usecase tests

type mockStockRepo struct {
	stocks map[int64]*domain.Stock
	err    error
}

func (m *mockStockRepo) Create(_ context.Context, stock *domain.Stock) error {
	if m.err != nil {
		return m.err
	}
	if m.stocks == nil {
		m.stocks = make(map[int64]*domain.Stock)
	}
	stock.ID = int64(len(m.stocks) + 1)
	m.stocks[stock.ID] = stock
	return nil
}

func (m *mockStockRepo) GetByID(_ context.Context, id int64, merchantID int64) (*domain.Stock, error) {
	if m.err != nil {
		return nil, m.err
	}
	s, ok := m.stocks[id]
	if !ok || s.MerchantID != merchantID {
		return nil, domain.ErrStockNotFound
	}
	return s, nil
}

func (m *mockStockRepo) List(_ context.Context, merchantID int64) ([]domain.Stock, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.Stock
	for _, s := range m.stocks {
		if s.MerchantID == merchantID {
			result = append(result, *s)
		}
	}
	if result == nil {
		return []domain.Stock{}, nil
	}
	return result, nil
}

func (m *mockStockRepo) ListChangedSince(_ context.Context, merchantID int64, since *time.Time) ([]domain.Stock, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.Stock
	for _, s := range m.stocks {
		if s.MerchantID != merchantID {
			continue
		}
		if since != nil && !s.UpdatedAt.After(*since) {
			continue
		}
		result = append(result, *s)
	}
	if result == nil {
		return []domain.Stock{}, nil
	}
	return result, nil
}

func (m *mockStockRepo) Update(_ context.Context, stock *domain.Stock) error {
	if m.err != nil {
		return m.err
	}
	existing, ok := m.stocks[stock.ID]
	if !ok || existing.MerchantID != stock.MerchantID {
		return domain.ErrStockNotFound
	}
	m.stocks[stock.ID] = stock
	return nil
}

func (m *mockStockRepo) Deduct(_ context.Context, merchantID, branchID int64, _ string, _ int64, items []repository.StockAdjustmentItem) error {
	if m.err != nil {
		return m.err
	}
	for _, it := range items {
		stock := m.find(merchantID, branchID, it.ProductItemID)
		if stock == nil {
			return domain.ErrStockNotFound
		}
		if stock.Available < it.Quantity {
			return domain.ErrInsufficientStock
		}
		stock.Available -= it.Quantity
	}
	return nil
}

func (m *mockStockRepo) Restore(_ context.Context, merchantID, branchID int64, _ string, _ int64, items []repository.StockAdjustmentItem) error {
	if m.err != nil {
		return m.err
	}
	for _, it := range items {
		stock := m.find(merchantID, branchID, it.ProductItemID)
		if stock == nil {
			return domain.ErrStockNotFound
		}
		stock.Available += it.Quantity
	}
	return nil
}

func (m *mockStockRepo) find(merchantID, branchID, productItemID int64) *domain.Stock {
	if m.stocks == nil {
		return nil
	}
	for _, s := range m.stocks {
		if s.MerchantID == merchantID && s.BranchID == branchID && s.ProductItemID == productItemID {
			return s
		}
	}
	return nil
}

type mockComponentRepo struct {
	components map[int64]*domain.ProductComponent
	itemSeq    int64
	seq        int64
	err        error
}

type mockUnitRepo struct {
	units map[int64]*domain.Unit
}

func newMockUnitRepo() *mockUnitRepo {
	return &mockUnitRepo{units: map[int64]*domain.Unit{
		1: {ID: 1, Name: "Piece", Symbol: "pc", FactorToBase: 1},
		3: {ID: 3, Name: "Kilogram", Symbol: "kg", FactorToBase: 1000},
		4: {ID: 4, Name: "Gram", Symbol: "g", FactorToBase: 1},
	}}
}

func (m *mockUnitRepo) GetByID(_ context.Context, id int64) (*domain.Unit, error) {
	if m.units == nil {
		return nil, nil
	}
	return m.units[id], nil
}

func (m *mockComponentRepo) Create(_ context.Context, component *domain.ProductComponent) error {
	if m.err != nil {
		return m.err
	}
	if m.components == nil {
		m.components = make(map[int64]*domain.ProductComponent)
	}
	for _, existing := range m.components {
		if existing.ProductItemID == component.ProductItemID && existing.MerchantID == component.MerchantID {
			return domain.ErrComponentAlreadyExists
		}
	}
	m.seq++
	component.ID = m.seq
	for i := range component.Items {
		m.itemSeq++
		component.Items[i].ID = m.itemSeq
		component.Items[i].ProductComponentID = component.ID
	}
	m.components[component.ID] = component
	return nil
}

func (m *mockComponentRepo) GetByProductItem(_ context.Context, productItemID int64, merchantID int64) (*domain.ProductComponent, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, c := range m.components {
		if c.ProductItemID == productItemID && c.MerchantID == merchantID {
			return c, nil
		}
	}
	return nil, domain.ErrProductComponentNotFound
}

func (m *mockComponentRepo) List(_ context.Context, merchantID int64) ([]domain.ProductComponent, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.ProductComponent
	for _, c := range m.components {
		if c.MerchantID == merchantID {
			result = append(result, *c)
		}
	}
	if result == nil {
		return []domain.ProductComponent{}, nil
	}
	return result, nil
}

func (m *mockComponentRepo) Update(_ context.Context, component *domain.ProductComponent) error {
	if m.err != nil {
		return m.err
	}
	existing, ok := m.components[component.ID]
	if !ok || existing.MerchantID != component.MerchantID {
		return domain.ErrProductComponentNotFound
	}
	m.components[component.ID] = component
	return nil
}

func (m *mockComponentRepo) Delete(_ context.Context, id int64, merchantID int64) error {
	if m.err != nil {
		return m.err
	}
	c, ok := m.components[id]
	if !ok || c.MerchantID != merchantID {
		return domain.ErrProductComponentNotFound
	}
	delete(m.components, id)
	return nil
}

// Test helpers

func int64Ptr(v int64) *int64 { return &v }
