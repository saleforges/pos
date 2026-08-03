package memory

import (
	"context"
	"sync"
	"time"

	"github.com/saleforge/pos/services/internal/inventory/domain"
	"github.com/saleforge/pos/services/internal/inventory/port/repository"
)

var _ repository.StockRepository = (*StockRepository)(nil)
var _ repository.StockAdjustmentRepository = (*StockRepository)(nil)

type StockRepository struct {
	mu        sync.RWMutex
	stocks    map[int64]*domain.Stock
	movements []domain.StockMovement
	seq       int64
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

func (r *StockRepository) ListChangedSince(_ context.Context, merchantID int64, since *time.Time) ([]domain.Stock, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Stock
	for _, s := range r.stocks {
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

// Deduct applies a batch of stock deductions atomically and records a
// stock_out movement per line. Fails the whole batch if any item has
// insufficient available stock.
func (r *StockRepository) Deduct(_ context.Context, merchantID, branchID int64, referenceType string, referenceID int64, items []repository.StockAdjustmentItem) error {
	return r.adjust(merchantID, branchID, referenceType, referenceID, items, domain.MovementTypeStockOut, -1)
}

// Restore applies a batch of stock restores atomically and records a
// stock_in movement per line (e.g. when an order is cancelled).
func (r *StockRepository) Restore(_ context.Context, merchantID, branchID int64, referenceType string, referenceID int64, items []repository.StockAdjustmentItem) error {
	return r.adjust(merchantID, branchID, referenceType, referenceID, items, domain.MovementTypeStockIn, 1)
}

func (r *StockRepository) adjust(merchantID, branchID int64, referenceType string, referenceID int64, items []repository.StockAdjustmentItem, movementType domain.MovementType, sign int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, it := range items {
		if it.ProductItemID == 0 || it.Quantity <= 0 {
			return domain.ErrInvalidStock
		}
		// find stock row for branch+item
		var stock *domain.Stock
		for _, s := range r.stocks {
			if s.MerchantID == merchantID && s.BranchID == branchID && s.ProductItemID == it.ProductItemID {
				stock = s
				break
			}
		}
		if stock == nil {
			return domain.ErrStockNotFound
		}
		if sign < 0 && stock.Available < it.Quantity {
			return domain.ErrInsufficientStock
		}
		stock.Available += sign * it.Quantity
		stock.UpdatedAt = time.Now().UTC()

		r.seq++
		r.movements = append(r.movements, domain.StockMovement{
			ID:            r.seq,
			MerchantID:    merchantID,
			BranchID:      branchID,
			ProductItemID: it.ProductItemID,
			Type:          movementType,
			Quantity:      it.Quantity,
			ReferenceType: referenceType,
			ReferenceID:   &referenceID,
			CreatedAt:     time.Now().UTC(),
		})
	}
	return nil
}
