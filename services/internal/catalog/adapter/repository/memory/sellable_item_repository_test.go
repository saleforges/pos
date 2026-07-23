package memory

import (
	"context"
	"testing"
	"time"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

func TestSellableItemRepository(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	now := time.Now().UTC()

	t.Run("create and get by id", func(t *testing.T) {
		repo := NewSellableItemRepository()
		item := &domain.SellableItem{ProductID: 1, Name: "Marning Curah", UnitID: 3, Price: 15000, TrackInventory: true, Status: domain.SellableItemStatusActive, CreatedAt: now, UpdatedAt: now}
		if err := repo.Create(ctx, item); err != nil {
			t.Fatalf("Create: %v", err)
		}
		if item.ID == 0 {
			t.Error("expected non-zero id")
		}
		got, err := repo.GetByID(ctx, item.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Name != "Marning Curah" {
			t.Errorf("expected 'Marning Curah', got '%s'", got.Name)
		}
	})

	t.Run("get by id not found", func(t *testing.T) {
		repo := NewSellableItemRepository()
		_, err := repo.GetByID(ctx, 999)
		if err != domain.ErrSellableItemNotFound {
			t.Errorf("expected ErrSellableItemNotFound, got %v", err)
		}
	})

	t.Run("list by product", func(t *testing.T) {
		repo := NewSellableItemRepository()
		repo.Create(ctx, &domain.SellableItem{ProductID: 1, Name: "Item A", UnitID: 1, Price: 10000, Status: domain.SellableItemStatusActive, CreatedAt: now, UpdatedAt: now})
		repo.Create(ctx, &domain.SellableItem{ProductID: 1, Name: "Item B", UnitID: 1, Price: 20000, Status: domain.SellableItemStatusActive, CreatedAt: now, UpdatedAt: now})
		repo.Create(ctx, &domain.SellableItem{ProductID: 2, Name: "Item C", UnitID: 1, Price: 30000, Status: domain.SellableItemStatusActive, CreatedAt: now, UpdatedAt: now})

		items, err := repo.ListByProduct(ctx, 1)
		if err != nil {
			t.Fatalf("ListByProduct: %v", err)
		}
		if len(items) != 2 {
			t.Errorf("expected 2 items, got %d", len(items))
		}
	})

	t.Run("update", func(t *testing.T) {
		repo := NewSellableItemRepository()
		repo.Create(ctx, &domain.SellableItem{ProductID: 1, Name: "Old", UnitID: 1, Price: 5000, Status: domain.SellableItemStatusActive, CreatedAt: now, UpdatedAt: now})
		item, _ := repo.GetByID(ctx, 1)
		item.Name = "New"
		if err := repo.Update(ctx, item); err != nil {
			t.Fatalf("Update: %v", err)
		}
		got, _ := repo.GetByID(ctx, 1)
		if got.Name != "New" {
			t.Errorf("expected 'New', got '%s'", got.Name)
		}
	})

	t.Run("delete", func(t *testing.T) {
		repo := NewSellableItemRepository()
		repo.Create(ctx, &domain.SellableItem{ProductID: 1, Name: "Del", UnitID: 1, Price: 10000, Status: domain.SellableItemStatusActive, CreatedAt: now, UpdatedAt: now})
		if err := repo.Delete(ctx, 1); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		_, err := repo.GetByID(ctx, 1)
		if err != domain.ErrSellableItemNotFound {
			t.Errorf("expected ErrSellableItemNotFound after delete, got %v", err)
		}
	})
}

func TestBarcodeRepository(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("create and list by sellable item", func(t *testing.T) {
		repo := NewBarcodeRepository()
		b := &domain.SellableItemBarcode{SellableItemID: 1, Barcode: "899999999"}
		if err := repo.Create(ctx, b); err != nil {
			t.Fatalf("Create: %v", err)
		}
		if b.ID == 0 {
			t.Error("expected non-zero id")
		}
		list, err := repo.ListBySellableItem(ctx, 1)
		if err != nil {
			t.Fatalf("ListBySellableItem: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 barcode, got %d", len(list))
		}
	})

	t.Run("duplicate barcode returns error", func(t *testing.T) {
		repo := NewBarcodeRepository()
		repo.Create(ctx, &domain.SellableItemBarcode{SellableItemID: 1, Barcode: "dup"})
		err := repo.Create(ctx, &domain.SellableItemBarcode{SellableItemID: 2, Barcode: "dup"})
		if err != domain.ErrBarcodeExists {
			t.Errorf("expected ErrBarcodeExists, got %v", err)
		}
	})

	t.Run("get by barcode", func(t *testing.T) {
		repo := NewBarcodeRepository()
		repo.Create(ctx, &domain.SellableItemBarcode{SellableItemID: 1, Barcode: "findme"})
		b, err := repo.GetByBarcode(ctx, "findme")
		if err != nil {
			t.Fatalf("GetByBarcode: %v", err)
		}
		if b.Barcode != "findme" {
			t.Errorf("expected 'findme', got '%s'", b.Barcode)
		}
	})
}
