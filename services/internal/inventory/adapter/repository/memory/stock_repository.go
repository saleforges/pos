package memory

import (
	"context"
	"sync"

	"github.com/saleforge/pos/services/internal/inventory/domain"
	"github.com/saleforge/pos/services/internal/inventory/port/repository"
)

var _ repository.StockRepository = (*StockRepository)(nil)

type StockRepository struct {
	mu    sync.RWMutex
	stocks map[int64]*domain.Stock
	seq   int64
}

func NewStockRepository() *StockRepository {
	return &StockRepository{
		stocks: make(map[int64]*domain.Stock),
	}
}

func (r *StockRepository) Create(_ context.Context, stock *domain.Stock) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.seq++
	stock.ID = r.seq
	r.stocks[stock.ID] = stock
	return nil
}

func (r *StockRepository) GetByID(_ context.Context, id int64, merchantID int64) (*domain.Stock, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.stocks[id]
	if !ok || s.MerchantID != merchantID {
		return nil, domain.ErrStockNotFound
	}
	return s, nil
}

func (r *StockRepository) List(_ context.Context, merchantID int64) ([]domain.Stock, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Stock
	for _, s := range r.stocks {
		if s.MerchantID == merchantID {
			result = append(result, *s)
		}
	}
	if result == nil {
		return []domain.Stock{}, nil
	}
	return result, nil
}

func (r *StockRepository) Update(_ context.Context, stock *domain.Stock) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.stocks[stock.ID]
	if !ok || existing.MerchantID != stock.MerchantID {
		return domain.ErrStockNotFound
	}
	r.stocks[stock.ID] = stock
	return nil
}
