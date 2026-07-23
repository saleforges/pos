package memory

import (
	"context"
	"sync"
	"time"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
)

var _ repository.SellableItemRepository = (*SellableItemRepository)(nil)
var _ repository.BarcodeRepository = (*BarcodeRepository)(nil)

type SellableItemRepository struct {
	mu    sync.RWMutex
	items map[int64]*domain.SellableItem
	seq   int64
}

func NewSellableItemRepository() *SellableItemRepository {
	return &SellableItemRepository{
		items: make(map[int64]*domain.SellableItem),
	}
}

func (r *SellableItemRepository) Create(_ context.Context, item *domain.SellableItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	item.ID = r.seq
	r.items[item.ID] = item
	return nil
}

func (r *SellableItemRepository) GetByID(_ context.Context, id int64) (*domain.SellableItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[id]
	if !ok || item.DeletedAt != nil {
		return nil, domain.ErrSellableItemNotFound
	}
	return item, nil
}

func (r *SellableItemRepository) ListByProduct(_ context.Context, productID int64) ([]domain.SellableItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.SellableItem
	for _, item := range r.items {
		if item.ProductID == productID && item.DeletedAt == nil {
			result = append(result, *item)
		}
	}
	if result == nil {
		return []domain.SellableItem{}, nil
	}
	return result, nil
}

func (r *SellableItemRepository) Update(_ context.Context, item *domain.SellableItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[item.ID] = item
	return nil
}

func (r *SellableItemRepository) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if item, ok := r.items[id]; ok {
		now := time.Now().UTC()
		item.DeletedAt = &now
	}
	return nil
}

func (r *SellableItemRepository) Restore(_ context.Context, id int64) (*domain.SellableItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.items[id]
	if !ok {
		return nil, domain.ErrSellableItemNotFound
	}
	item.DeletedAt = nil
	return item, nil
}

type BarcodeRepository struct {
	mu       sync.RWMutex
	barcodes map[int64]*domain.SellableItemBarcode
	seq      int64
}

func NewBarcodeRepository() *BarcodeRepository {
	return &BarcodeRepository{
		barcodes: make(map[int64]*domain.SellableItemBarcode),
	}
}

func (r *BarcodeRepository) Create(_ context.Context, barcode *domain.SellableItemBarcode) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, b := range r.barcodes {
		if b.Barcode == barcode.Barcode {
			return domain.ErrBarcodeExists
		}
	}
	r.seq++
	barcode.ID = r.seq
	r.barcodes[barcode.ID] = barcode
	return nil
}

func (r *BarcodeRepository) GetByBarcode(_ context.Context, barcode string) (*domain.SellableItemBarcode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, b := range r.barcodes {
		if b.Barcode == barcode {
			return b, nil
		}
	}
	return nil, domain.ErrSellableItemNotFound
}

func (r *BarcodeRepository) ListBySellableItem(_ context.Context, sellableItemID int64) ([]domain.SellableItemBarcode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.SellableItemBarcode
	for _, b := range r.barcodes {
		if b.SellableItemID == sellableItemID {
			result = append(result, *b)
		}
	}
	if result == nil {
		return []domain.SellableItemBarcode{}, nil
	}
	return result, nil
}

func (r *BarcodeRepository) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.barcodes, id)
	return nil
}
