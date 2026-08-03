package memory

import (
	"context"
	"sync"

	"github.com/saleforge/pos/services/internal/inventory/domain"
	"github.com/saleforge/pos/services/internal/inventory/port/repository"
)

var _ repository.StockMovementRepository = (*StockMovementRepository)(nil)

type StockMovementRepository struct {
	mu        sync.RWMutex
	movements map[int64]*domain.StockMovement
	seq       int64
}

func NewStockMovementRepository() *StockMovementRepository {
	return &StockMovementRepository{
		movements: make(map[int64]*domain.StockMovement),
	}
}

func (r *StockMovementRepository) Create(_ context.Context, m *domain.StockMovement) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.seq++
	m.ID = r.seq
	r.movements[m.ID] = m
	return nil
}

func (r *StockMovementRepository) ListByProductItem(_ context.Context, productItemID int64, merchantID int64) ([]domain.StockMovement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.StockMovement
	for _, m := range r.movements {
		if m.ProductItemID == productItemID && m.MerchantID == merchantID {
			result = append(result, *m)
		}
	}
	if result == nil {
		return []domain.StockMovement{}, nil
	}
	return result, nil
}
