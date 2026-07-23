package usecase

import (
	"context"
	"testing"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

func TestProductItemUsecase_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "Marning"})
		mockUnit := newMockUnitRepo()
		uc := NewProductItemUsecase(&mockProductItemRepo{}, mockProd, mockUnit)

		item, err := uc.Create(ctx, CreateProductItemParams{
			MerchantID: 1, ProductID: 1, Name: "Marning Curah", PriceAmount: 15000, UnitID: int64Ptr(2), TrackInventory: true,
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
		if item.Price.Amount != 15000 {
			t.Errorf("expected price 15000, got %f", item.Price.Amount)
		}
		if item.Price.Currency != "IDR" {
			t.Errorf("expected currency IDR, got %s", item.Price.Currency)
		}
		if !item.TrackInventory {
			t.Error("expected track_inventory=true")
		}
		if item.MerchantID != 1 {
			t.Errorf("expected merchantID 1, got %d", item.MerchantID)
		}
	})

	t.Run("empty name returns error", func(t *testing.T) {
		uc := NewProductItemUsecase(&mockProductItemRepo{}, &mockProductRepo{}, newMockUnitRepo())
		_, err := uc.Create(ctx, CreateProductItemParams{MerchantID: 1, ProductID: 1})
		if err != domain.ErrInvalidProductItem {
			t.Errorf("expected ErrInvalidProductItem, got %v", err)
		}
	})

	t.Run("invalid product returns error", func(t *testing.T) {
		uc := NewProductItemUsecase(&mockProductItemRepo{}, &mockProductRepo{}, newMockUnitRepo())
		_, err := uc.Create(ctx, CreateProductItemParams{MerchantID: 1, ProductID: 999, Name: "Item"})
		if err != domain.ErrProductNotFound {
			t.Errorf("expected ErrProductNotFound, got %v", err)
		}
	})

	t.Run("merchant cannot create item for another merchant's product", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P1"})
		uc := NewProductItemUsecase(&mockProductItemRepo{}, mockProd, newMockUnitRepo())

		_, err := uc.Create(ctx, CreateProductItemParams{
			MerchantID: 2, ProductID: 1, Name: "Item",
		})
		if err != domain.ErrProductNotFound {
			t.Errorf("expected ErrProductNotFound for cross-merchant, got %v", err)
		}
	})

	t.Run("invalid unit returns error", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P"})
		uc := NewProductItemUsecase(&mockProductItemRepo{}, mockProd, newMockUnitRepo())
		_, err := uc.Create(ctx, CreateProductItemParams{MerchantID: 1, ProductID: 1, Name: "Item", UnitID: int64Ptr(999)})
		if err != domain.ErrUnitNotFound {
			t.Errorf("expected ErrUnitNotFound, got %v", err)
		}
	})

	t.Run("nullable unit", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "Es Teh"})
		uc := NewProductItemUsecase(&mockProductItemRepo{}, mockProd, newMockUnitRepo())

		item, err := uc.Create(ctx, CreateProductItemParams{
			MerchantID: 1, ProductID: 1, Name: "Es Teh", PriceAmount: 3000,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item.UnitID != nil {
			t.Error("expected nil unit for simple product")
		}
	})

	t.Run("sku is optional", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "Marning"})
		uc := NewProductItemUsecase(&mockProductItemRepo{}, mockProd, newMockUnitRepo())

		item, err := uc.Create(ctx, CreateProductItemParams{
			MerchantID: 1, ProductID: 1, Name: "Marning Pack", SKU: "", PriceAmount: 1000,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item.SKU != "" {
			t.Errorf("expected empty sku, got '%s'", item.SKU)
		}
	})

	// SKU uniqueness tests
	t.Run("same merchant + same SKU => rejected", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P1"})
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P2"})
		mockItem := &mockProductItemRepo{}
		uc := NewProductItemUsecase(mockItem, mockProd, newMockUnitRepo())

		// Create first item with SKU
		_, err := uc.Create(ctx, CreateProductItemParams{
			MerchantID: 1, ProductID: 1, Name: "Item A", SKU: "SKU-001", PriceAmount: 1000,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Same merchant, same SKU => rejected
		_, err = uc.Create(ctx, CreateProductItemParams{
			MerchantID: 1, ProductID: 2, Name: "Item B", SKU: "SKU-001", PriceAmount: 2000,
		})
		if err != domain.ErrSKUDuplicate {
			t.Errorf("expected ErrSKUDuplicate for duplicate SKU within same merchant, got %v", err)
		}
	})

	t.Run("same merchant + different SKU => allowed", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P1"})
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P2"})
		mockItem := &mockProductItemRepo{}
		uc := NewProductItemUsecase(mockItem, mockProd, newMockUnitRepo())

		_, err := uc.Create(ctx, CreateProductItemParams{
			MerchantID: 1, ProductID: 1, Name: "Item A", SKU: "SKU-001", PriceAmount: 1000,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Same merchant, different SKU => allowed
		_, err = uc.Create(ctx, CreateProductItemParams{
			MerchantID: 1, ProductID: 2, Name: "Item B", SKU: "SKU-002", PriceAmount: 2000,
		})
		if err != nil {
			t.Fatalf("expected success for different SKU, got %v", err)
		}
	})

	t.Run("different merchant + same SKU => allowed", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P1"})
		mockProd.Create(ctx, &domain.Product{MerchantID: 2, CategoryID: 1, Name: "P2"})
		mockItem := &mockProductItemRepo{}
		uc := NewProductItemUsecase(mockItem, mockProd, newMockUnitRepo())

		// Merchant 1 creates with SKU
		_, err := uc.Create(ctx, CreateProductItemParams{
			MerchantID: 1, ProductID: 1, Name: "Item A", SKU: "SKU-001", PriceAmount: 1000,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Merchant 2 can use same SKU
		_, err = uc.Create(ctx, CreateProductItemParams{
			MerchantID: 2, ProductID: 2, Name: "Item B", SKU: "SKU-001", PriceAmount: 2000,
		})
		if err != nil {
			t.Fatalf("expected success for different merchant same SKU, got %v", err)
		}
	})
}

func TestProductItemUsecase_ListByProduct(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("lists items for product", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P"})
		mockItem := &mockProductItemRepo{}
		mockItem.Create(ctx, &domain.ProductItem{ProductID: 1, MerchantID: 1, Name: "Item A"})
		mockItem.Create(ctx, &domain.ProductItem{ProductID: 1, MerchantID: 1, Name: "Item B"})

		uc := NewProductItemUsecase(mockItem, mockProd, newMockUnitRepo())
		items, err := uc.ListByProduct(ctx, 1, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 2 {
			t.Errorf("expected 2 items, got %d", len(items))
		}
	})

	t.Run("invalid product returns error", func(t *testing.T) {
		uc := NewProductItemUsecase(&mockProductItemRepo{}, &mockProductRepo{}, newMockUnitRepo())
		_, err := uc.ListByProduct(ctx, 999, 1)
		if err != domain.ErrProductNotFound {
			t.Errorf("expected ErrProductNotFound, got %v", err)
		}
	})

	t.Run("merchant A cannot list items of merchant B product", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P1"})
		uc := NewProductItemUsecase(&mockProductItemRepo{}, mockProd, newMockUnitRepo())

		_, err := uc.ListByProduct(ctx, 1, 2)
		if err != domain.ErrProductNotFound {
			t.Errorf("expected ErrProductNotFound for cross-merchant, got %v", err)
		}
	})
}

func TestProductItemUsecase_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("partial update", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P"})
		mockItem := &mockProductItemRepo{}
		mockItem.Create(ctx, &domain.ProductItem{ProductID: 1, MerchantID: 1, Name: "Old", Price: domain.Price{Amount: 5000, Currency: "IDR"}})
		uc := NewProductItemUsecase(mockItem, mockProd, newMockUnitRepo())

		name := "New"
		item, err := uc.Update(ctx, UpdateProductItemParams{ID: 1, MerchantID: 1, Name: &name})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item.Name != "New" {
			t.Errorf("expected 'New', got '%s'", item.Name)
		}
	})

	t.Run("update sku", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P"})
		mockItem := &mockProductItemRepo{}
		mockItem.Create(ctx, &domain.ProductItem{ProductID: 1, MerchantID: 1, Name: "Item", Price: domain.Price{Amount: 1000, Currency: "IDR"}})
		uc := NewProductItemUsecase(mockItem, mockProd, newMockUnitRepo())

		sku := "NEW-SKU-001"
		item, err := uc.Update(ctx, UpdateProductItemParams{ID: 1, MerchantID: 1, SKU: &sku})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if item.SKU != "NEW-SKU-001" {
			t.Errorf("expected 'NEW-SKU-001', got '%s'", item.SKU)
		}
	})

	t.Run("merchant cannot update another merchant's item", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P"})
		mockItem := &mockProductItemRepo{}
		mockItem.Create(ctx, &domain.ProductItem{ProductID: 1, MerchantID: 1, Name: "Item", Price: domain.Price{Amount: 1000, Currency: "IDR"}})
		uc := NewProductItemUsecase(mockItem, mockProd, newMockUnitRepo())

		name := "Hacked"
		_, err := uc.Update(ctx, UpdateProductItemParams{ID: 1, MerchantID: 2, Name: &name})
		if err != domain.ErrProductItemNotFound {
			t.Errorf("expected ErrProductItemNotFound for cross-merchant update, got %v", err)
		}
	})

	t.Run("sku uniqueness enforced on update", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P1"})
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P2"})
		mockItem := &mockProductItemRepo{}
		mockItem.Create(ctx, &domain.ProductItem{ProductID: 1, MerchantID: 1, Name: "Item A", SKU: "SKU-001", Price: domain.Price{Amount: 1000, Currency: "IDR"}})
		mockItem.Create(ctx, &domain.ProductItem{ProductID: 2, MerchantID: 1, Name: "Item B", SKU: "SKU-002", Price: domain.Price{Amount: 2000, Currency: "IDR"}})
		uc := NewProductItemUsecase(mockItem, mockProd, newMockUnitRepo())

		// Try to change Item B's SKU to Item A's SKU
		sku := "SKU-001"
		_, err := uc.Update(ctx, UpdateProductItemParams{ID: 2, MerchantID: 1, SKU: &sku})
		if err != domain.ErrSKUDuplicate {
			t.Errorf("expected ErrSKUDuplicate for duplicate SKU on update, got %v", err)
		}
	})
}

func TestProductItemUsecase_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("merchant cannot delete another merchant's item", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P"})
		mockItem := &mockProductItemRepo{}
		mockItem.Create(ctx, &domain.ProductItem{ProductID: 1, MerchantID: 1, Name: "Item", Price: domain.Price{Amount: 1000, Currency: "IDR"}})
		uc := NewProductItemUsecase(mockItem, mockProd, newMockUnitRepo())

		err := uc.Delete(ctx, 1, 2)
		if err != domain.ErrProductItemNotFound {
			t.Errorf("expected ErrProductItemNotFound for cross-merchant delete, got %v", err)
		}
	})

	t.Run("own item delete succeeds", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P"})
		mockItem := &mockProductItemRepo{}
		mockItem.Create(ctx, &domain.ProductItem{ProductID: 1, MerchantID: 1, Name: "Item", Price: domain.Price{Amount: 1000, Currency: "IDR"}})
		uc := NewProductItemUsecase(mockItem, mockProd, newMockUnitRepo())

		err := uc.Delete(ctx, 1, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func int64Ptr(v int64) *int64 { return &v }
