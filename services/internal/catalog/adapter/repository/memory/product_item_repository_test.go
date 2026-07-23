package memory

import (
	"context"
	"testing"
	"time"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

func TestProductItemRepository(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	now := time.Now().UTC()

	t.Run("create and get by id", func(t *testing.T) {
		repo := NewProductItemRepository()
		item := &domain.ProductItem{
			ProductID:      1,
			MerchantID:     1,
			Name:           "Marning Curah",
			UnitID:         int64Ptr(3),
			Price:          domain.Price{Amount: 15000, Currency: "IDR"},
			TrackInventory: true,
			Status:         domain.ProductItemStatusActive,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := repo.Create(ctx, item); err != nil {
			t.Fatalf("Create: %v", err)
		}
		if item.ID == 0 {
			t.Error("expected non-zero id")
		}
		got, err := repo.GetByID(ctx, item.ID, 1)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Name != "Marning Curah" {
			t.Errorf("expected 'Marning Curah', got '%s'", got.Name)
		}
		if got.Price.Amount != 15000 {
			t.Errorf("expected price 15000, got %f", got.Price.Amount)
		}
	})

	t.Run("get by id different merchant returns not found", func(t *testing.T) {
		repo := NewProductItemRepository()
		item := &domain.ProductItem{ProductID: 1, MerchantID: 1, Name: "Item", Price: domain.Price{Amount: 1000, Currency: "IDR"}, Status: domain.ProductItemStatusActive, CreatedAt: now, UpdatedAt: now}
		repo.Create(ctx, item)
		_, err := repo.GetByID(ctx, item.ID, 2)
		if err != domain.ErrProductItemNotFound {
			t.Errorf("expected ErrProductItemNotFound for cross-merchant, got %v", err)
		}
	})

	t.Run("get by id not found", func(t *testing.T) {
		repo := NewProductItemRepository()
		_, err := repo.GetByID(ctx, 999, 1)
		if err != domain.ErrProductItemNotFound {
			t.Errorf("expected ErrProductItemNotFound, got %v", err)
		}
	})

	t.Run("list by product", func(t *testing.T) {
		repo := NewProductItemRepository()
		repo.Create(ctx, &domain.ProductItem{ProductID: 1, MerchantID: 1, Name: "Item A", Price: domain.Price{Amount: 10000, Currency: "IDR"}, Status: domain.ProductItemStatusActive, CreatedAt: now, UpdatedAt: now})
		repo.Create(ctx, &domain.ProductItem{ProductID: 1, MerchantID: 1, Name: "Item B", Price: domain.Price{Amount: 20000, Currency: "IDR"}, Status: domain.ProductItemStatusActive, CreatedAt: now, UpdatedAt: now})
		repo.Create(ctx, &domain.ProductItem{ProductID: 2, MerchantID: 2, Name: "Item C", Price: domain.Price{Amount: 30000, Currency: "IDR"}, Status: domain.ProductItemStatusActive, CreatedAt: now, UpdatedAt: now})

		items, err := repo.ListByProduct(ctx, 1, 1)
		if err != nil {
			t.Fatalf("ListByProduct: %v", err)
		}
		if len(items) != 2 {
			t.Errorf("expected 2 items, got %d", len(items))
		}
	})

	t.Run("list by product respects merchant scope", func(t *testing.T) {
		repo := NewProductItemRepository()
		repo.Create(ctx, &domain.ProductItem{ProductID: 1, MerchantID: 2, Name: "Other", Price: domain.Price{Amount: 1000, Currency: "IDR"}, Status: domain.ProductItemStatusActive, CreatedAt: now, UpdatedAt: now})
		items, err := repo.ListByProduct(ctx, 1, 1)
		if err != nil {
			t.Fatalf("ListByProduct: %v", err)
		}
		if len(items) != 0 {
			t.Errorf("expected 0 items for different merchant, got %d", len(items))
		}
	})

	t.Run("list by merchant", func(t *testing.T) {
		repo := NewProductItemRepository()
		repo.Create(ctx, &domain.ProductItem{ProductID: 1, MerchantID: 1, Name: "Item A", Price: domain.Price{Amount: 1000, Currency: "IDR"}, Status: domain.ProductItemStatusActive, CreatedAt: now, UpdatedAt: now})
		repo.Create(ctx, &domain.ProductItem{ProductID: 2, MerchantID: 2, Name: "Item B", Price: domain.Price{Amount: 2000, Currency: "IDR"}, Status: domain.ProductItemStatusActive, CreatedAt: now, UpdatedAt: now})
		items, err := repo.ListByMerchant(ctx, 1)
		if err != nil {
			t.Fatalf("ListByMerchant: %v", err)
		}
		if len(items) != 1 {
			t.Errorf("expected 1 item for merchant 1, got %d", len(items))
		}
	})

	t.Run("update", func(t *testing.T) {
		repo := NewProductItemRepository()
		repo.Create(ctx, &domain.ProductItem{ProductID: 1, MerchantID: 1, Name: "Old", Price: domain.Price{Amount: 5000, Currency: "IDR"}, Status: domain.ProductItemStatusActive, CreatedAt: now, UpdatedAt: now})
		item, _ := repo.GetByID(ctx, 1, 1)
		item.Name = "New"
		if err := repo.Update(ctx, item); err != nil {
			t.Fatalf("Update: %v", err)
		}
		got, _ := repo.GetByID(ctx, 1, 1)
		if got.Name != "New" {
			t.Errorf("expected 'New', got '%s'", got.Name)
		}
	})

	t.Run("cross-merchant update blocked", func(t *testing.T) {
		repo := NewProductItemRepository()
		repo.Create(ctx, &domain.ProductItem{ProductID: 1, MerchantID: 1, Name: "Orig", Price: domain.Price{Amount: 1000, Currency: "IDR"}, Status: domain.ProductItemStatusActive, CreatedAt: now, UpdatedAt: now})
		item := &domain.ProductItem{ID: 1, ProductID: 1, MerchantID: 2, Name: "Hacked", Price: domain.Price{Amount: 1000, Currency: "IDR"}, Status: domain.ProductItemStatusActive, CreatedAt: now, UpdatedAt: now}
		err := repo.Update(ctx, item)
		if err != domain.ErrProductItemNotFound {
			t.Errorf("expected ErrProductItemNotFound for cross-merchant update, got %v", err)
		}
	})

	t.Run("delete", func(t *testing.T) {
		repo := NewProductItemRepository()
		repo.Create(ctx, &domain.ProductItem{ProductID: 1, MerchantID: 1, Name: "Del", Price: domain.Price{Amount: 10000, Currency: "IDR"}, Status: domain.ProductItemStatusActive, CreatedAt: now, UpdatedAt: now})
		if err := repo.Delete(ctx, 1, 1); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		_, err := repo.GetByID(ctx, 1, 1)
		if err != domain.ErrProductItemNotFound {
			t.Errorf("expected ErrProductItemNotFound after delete, got %v", err)
		}
	})

	t.Run("cross-merchant delete blocked", func(t *testing.T) {
		repo := NewProductItemRepository()
		repo.Create(ctx, &domain.ProductItem{ProductID: 1, MerchantID: 1, Name: "Del", Price: domain.Price{Amount: 10000, Currency: "IDR"}, Status: domain.ProductItemStatusActive, CreatedAt: now, UpdatedAt: now})
		err := repo.Delete(ctx, 1, 2)
		if err != nil {
			t.Fatalf("Delete cross-merchant should not error but item should remain: %v", err)
		}
		// Item should still be accessible by its own merchant
		_, err = repo.GetByID(ctx, 1, 1)
		if err != nil {
			t.Errorf("expected item still accessible by own merchant, got %v", err)
		}
	})

	t.Run("sku uniqueness within merchant", func(t *testing.T) {
		repo := NewProductItemRepository()
		item1 := &domain.ProductItem{
			ProductID:  1,
			MerchantID: 1,
			Name:       "Marning Pack",
			SKU:        "MRN-PACK-001",
			Price:      domain.Price{Amount: 1000, Currency: "IDR"},
			Status:     domain.ProductItemStatusActive,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := repo.Create(ctx, item1); err != nil {
			t.Fatalf("Create: %v", err)
		}
		got, _ := repo.GetByID(ctx, item1.ID, 1)
		if got.SKU != "MRN-PACK-001" {
			t.Errorf("expected SKU 'MRN-PACK-001', got '%s'", got.SKU)
		}

		// Same merchant, same SKU => rejected
		item2 := &domain.ProductItem{
			ProductID:  2,
			MerchantID: 1,
			Name:       "Duplicate SKU",
			SKU:        "MRN-PACK-001",
			Price:      domain.Price{Amount: 2000, Currency: "IDR"},
			Status:     domain.ProductItemStatusActive,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		err := repo.Create(ctx, item2)
		if err != domain.ErrSKUDuplicate {
			t.Errorf("expected ErrSKUDuplicate, got %v", err)
		}

		// Different merchant, same SKU => allowed
		item3 := &domain.ProductItem{
			ProductID:  3,
			MerchantID: 2,
			Name:       "Other merchant product",
			SKU:        "MRN-PACK-001",
			Price:      domain.Price{Amount: 3000, Currency: "IDR"},
			Status:     domain.ProductItemStatusActive,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := repo.Create(ctx, item3); err != nil {
			t.Fatalf("expected success for different merchant same SKU, got %v", err)
		}
	})
}

func TestProductItemBarcodeRepository(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("create and list by product item", func(t *testing.T) {
		repo := NewProductItemBarcodeRepository()
		b := &domain.ProductItemBarcode{ProductItemID: 1, Barcode: "899999999"}
		if err := repo.Create(ctx, b); err != nil {
			t.Fatalf("Create: %v", err)
		}
		if b.ID == 0 {
			t.Error("expected non-zero id")
		}
		list, err := repo.ListByProductItem(ctx, 1)
		if err != nil {
			t.Fatalf("ListByProductItem: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 barcode, got %d", len(list))
		}
	})

	t.Run("duplicate barcode returns error", func(t *testing.T) {
		repo := NewProductItemBarcodeRepository()
		repo.Create(ctx, &domain.ProductItemBarcode{ProductItemID: 1, Barcode: "dup"})
		err := repo.Create(ctx, &domain.ProductItemBarcode{ProductItemID: 2, Barcode: "dup"})
		if err != domain.ErrBarcodeExists {
			t.Errorf("expected ErrBarcodeExists, got %v", err)
		}
	})

	t.Run("get by barcode", func(t *testing.T) {
		repo := NewProductItemBarcodeRepository()
		repo.Create(ctx, &domain.ProductItemBarcode{ProductItemID: 1, Barcode: "findme"})
		b, err := repo.GetByBarcode(ctx, "findme")
		if err != nil {
			t.Fatalf("GetByBarcode: %v", err)
		}
		if b.Barcode != "findme" {
			t.Errorf("expected 'findme', got '%s'", b.Barcode)
		}
	})
}

func int64Ptr(v int64) *int64 { return &v }
