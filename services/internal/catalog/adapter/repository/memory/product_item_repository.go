package memory

import (
	"context"
	"sync"
	"time"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
)

var _ repository.ProductItemRepository = (*ProductItemRepository)(nil)
var _ repository.ProductItemBarcodeRepository = (*ProductItemBarcodeRepository)(nil)

type ProductItemRepository struct {
	mu    sync.RWMutex
	items map[int64]*domain.ProductItem
	seq   int64
}

func NewProductItemRepository() *ProductItemRepository {
	return &ProductItemRepository{
		items: make(map[int64]*domain.ProductItem),
	}
}

func (r *ProductItemRepository) Create(_ context.Context, item *domain.ProductItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check SKU uniqueness within merchant scope
	if item.SKU != "" {
		for _, existing := range r.items {
			if existing.SKU == item.SKU && existing.MerchantID == item.MerchantID && existing.DeletedAt == nil {
				return domain.ErrSKUDuplicate
			}
		}
	}

	r.seq++
	item.ID = r.seq
	r.items[item.ID] = item
	return nil
}

func (r *ProductItemRepository) GetByID(_ context.Context, id int64, merchantID int64) (*domain.ProductItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[id]
	if !ok || item.DeletedAt != nil || item.MerchantID != merchantID {
		return nil, domain.ErrProductItemNotFound
	}
	return item, nil
}

func (r *ProductItemRepository) ListByProduct(_ context.Context, productID int64, merchantID int64) ([]domain.ProductItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.ProductItem
	for _, item := range r.items {
		if item.ProductID == productID && item.MerchantID == merchantID && item.DeletedAt == nil {
			result = append(result, *item)
		}
	}
	if result == nil {
		return []domain.ProductItem{}, nil
	}
	return result, nil
}

func (r *ProductItemRepository) ListByMerchant(_ context.Context, merchantID int64) ([]domain.ProductItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.ProductItem
	for _, item := range r.items {
		if item.MerchantID == merchantID && item.DeletedAt == nil {
			result = append(result, *item)
		}
	}
	if result == nil {
		return []domain.ProductItem{}, nil
	}
	return result, nil
}

func (r *ProductItemRepository) Update(_ context.Context, item *domain.ProductItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.items[item.ID]
	if !ok || existing.MerchantID != item.MerchantID {
		return domain.ErrProductItemNotFound
	}

	// Check SKU uniqueness within merchant scope (skip if same item)
	if item.SKU != "" {
		for _, other := range r.items {
			if other.ID != item.ID && other.SKU == item.SKU && other.MerchantID == item.MerchantID && other.DeletedAt == nil {
				return domain.ErrSKUDuplicate
			}
		}
	}

	r.items[item.ID] = item
	return nil
}

func (r *ProductItemRepository) Delete(_ context.Context, id int64, merchantID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if item, ok := r.items[id]; ok && item.MerchantID == merchantID {
		now := time.Now().UTC()
		item.DeletedAt = &now
	}
	return nil
}

func (r *ProductItemRepository) Restore(_ context.Context, id int64, merchantID int64) (*domain.ProductItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.items[id]
	if !ok || item.MerchantID != merchantID {
		return nil, domain.ErrProductItemNotFound
	}
	item.DeletedAt = nil
	return item, nil
}

func (r *ProductItemRepository) ListUpdatedAfter(_ context.Context, merchantID int64, after time.Time) ([]domain.ProductItem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.ProductItem
	for _, item := range r.items {
		if item.MerchantID != merchantID {
			continue
		}
		if item.UpdatedAt.After(after) || (item.DeletedAt != nil && item.DeletedAt.After(after)) {
			result = append(result, *item)
		}
	}
	if result == nil {
		return []domain.ProductItem{}, nil
	}
	return result, nil
}

type ProductItemBarcodeRepository struct {
	mu        sync.RWMutex
	barcodes  map[int64]*domain.ProductItemBarcode
	seq       int64
}

func NewProductItemBarcodeRepository() *ProductItemBarcodeRepository {
	return &ProductItemBarcodeRepository{
		barcodes: make(map[int64]*domain.ProductItemBarcode),
	}
}

func (r *ProductItemBarcodeRepository) Create(_ context.Context, barcode *domain.ProductItemBarcode) error {
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

func (r *ProductItemBarcodeRepository) GetByBarcode(_ context.Context, barcode string) (*domain.ProductItemBarcode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, b := range r.barcodes {
		if b.Barcode == barcode {
			return b, nil
		}
	}
	return nil, domain.ErrProductItemNotFound
}

func (r *ProductItemBarcodeRepository) ListByProductItem(_ context.Context, productItemID int64) ([]domain.ProductItemBarcode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.ProductItemBarcode
	for _, b := range r.barcodes {
		if b.ProductItemID == productItemID {
			result = append(result, *b)
		}
	}
	if result == nil {
		return []domain.ProductItemBarcode{}, nil
	}
	return result, nil
}

func (r *ProductItemBarcodeRepository) Delete(_ context.Context, id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.barcodes, id)
	return nil
}

func (r *ProductItemBarcodeRepository) ListByMerchant(_ context.Context, merchantID int64) ([]domain.ProductItemBarcode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []domain.ProductItemBarcode
	for _, b := range r.barcodes {
		result = append(result, *b)
	}
	if result == nil {
		return []domain.ProductItemBarcode{}, nil
	}
	return result, nil
}
