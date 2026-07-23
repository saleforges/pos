package usecase

import (
	"context"
	"testing"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

func TestSellableItemUsecase_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "Marning"})
		mockUnit := newMockUnitRepo()
		uc := NewSellableItemUsecase(&mockSellableItemRepo{}, mockProd, mockUnit)

		item, err := uc.Create(ctx, CreateSellableItemParams{
			ProductID: 1, Name: "Marning Curah", UnitID: 1, TrackInventory: true,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item.ID == 0 {
			t.Error("expected non-zero id")
		}
		if item.Name != "Marning Curah" {
			t.Errorf("expected 'Marning Curah', got '%s'", item.Name)
		}
		if !item.TrackInventory {
			t.Error("expected track_inventory=true")
		}
	})

	t.Run("empty name returns error", func(t *testing.T) {
		uc := NewSellableItemUsecase(&mockSellableItemRepo{}, &mockProductRepo{}, newMockUnitRepo())
		_, err := uc.Create(ctx, CreateSellableItemParams{ProductID: 1, UnitID: 1})
		if err != domain.ErrInvalidSellableItem {
			t.Errorf("expected ErrInvalidSellableItem, got %v", err)
		}
	})

	t.Run("invalid product returns error", func(t *testing.T) {
		uc := NewSellableItemUsecase(&mockSellableItemRepo{}, &mockProductRepo{}, newMockUnitRepo())
		_, err := uc.Create(ctx, CreateSellableItemParams{ProductID: 999, Name: "Item", UnitID: 1})
		if err != domain.ErrProductNotFound {
			t.Errorf("expected ErrProductNotFound, got %v", err)
		}
	})

	t.Run("invalid unit returns error", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P"})
		uc := NewSellableItemUsecase(&mockSellableItemRepo{}, mockProd, newMockUnitRepo())
		_, err := uc.Create(ctx, CreateSellableItemParams{ProductID: 1, Name: "Item", UnitID: 999})
		if err != domain.ErrUnitNotFound {
			t.Errorf("expected ErrUnitNotFound, got %v", err)
		}
	})
}

func TestSellableItemUsecase_ListByProduct(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("lists items for product", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P"})
		mockItem := &mockSellableItemRepo{}
		mockItem.Create(ctx, &domain.SellableItem{ProductID: 1, Name: "Item A", UnitID: 1})
		mockItem.Create(ctx, &domain.SellableItem{ProductID: 1, Name: "Item B", UnitID: 1})

		uc := NewSellableItemUsecase(mockItem, mockProd, newMockUnitRepo())
		items, err := uc.ListByProduct(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 2 {
			t.Errorf("expected 2 items, got %d", len(items))
		}
	})

	t.Run("invalid product returns error", func(t *testing.T) {
		uc := NewSellableItemUsecase(&mockSellableItemRepo{}, &mockProductRepo{}, newMockUnitRepo())
		_, err := uc.ListByProduct(ctx, 999)
		if err != domain.ErrProductNotFound {
			t.Errorf("expected ErrProductNotFound, got %v", err)
		}
	})
}

func TestSellableItemUsecase_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("partial update", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P"})
		mockItem := &mockSellableItemRepo{}
		mockItem.Create(ctx, &domain.SellableItem{ProductID: 1, Name: "Old", UnitID: 1})
		uc := NewSellableItemUsecase(mockItem, mockProd, newMockUnitRepo())

		name := "New"
		item, err := uc.Update(ctx, UpdateSellableItemParams{ID: 1, Name: &name})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item.Name != "New" {
			t.Errorf("expected 'New', got '%s'", item.Name)
		}
	})
}
