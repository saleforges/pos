package usecase

import (
	"context"
	"testing"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/pkg/pagination"
)

func TestProductUsecase_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockCat := &mockCategoryRepo{categories: map[int64]*domain.Category{1: {ID: 1, Name: "Snack"}}}
		uc := NewProductUsecase(&mockProductRepo{}, mockCat, newMockUnitRepo())

		p, err := uc.Create(ctx, CreateProductParams{
			MerchantID: 1, CategoryID: 1, Name: "Marning",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.ID == 0 {
			t.Error("expected non-zero id")
		}
		if p.Name != "Marning" {
			t.Errorf("expected 'Marning', got '%s'", p.Name)
		}
		if p.MerchantID != 1 {
			t.Errorf("expected merchant 1, got %d", p.MerchantID)
		}
		if p.Status != domain.ProductStatusActive {
			t.Errorf("expected active status")
		}
	})

	t.Run("empty name returns error", func(t *testing.T) {
		uc := NewProductUsecase(&mockProductRepo{}, &mockCategoryRepo{}, newMockUnitRepo())
		_, err := uc.Create(ctx, CreateProductParams{MerchantID: 1, CategoryID: 1})
		if err != domain.ErrInvalidProduct {
			t.Errorf("expected ErrInvalidProduct, got %v", err)
		}
	})

	t.Run("invalid category returns error", func(t *testing.T) {
		uc := NewProductUsecase(&mockProductRepo{}, &mockCategoryRepo{}, newMockUnitRepo())
		_, err := uc.Create(ctx, CreateProductParams{
			MerchantID: 1, CategoryID: 999, Name: "Test",
		})
		if err != domain.ErrCategoryNotFound {
			t.Errorf("expected ErrCategoryNotFound, got %v", err)
		}
	})
}

func TestProductUsecase_GetByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("found", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P1"})
		uc := NewProductUsecase(mockProd, &mockCategoryRepo{}, newMockUnitRepo())

		p, err := uc.GetByID(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name != "P1" {
			t.Errorf("expected 'P1', got '%s'", p.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		uc := NewProductUsecase(&mockProductRepo{}, &mockCategoryRepo{}, newMockUnitRepo())
		_, err := uc.GetByID(ctx, 999)
		if err != domain.ErrProductNotFound {
			t.Errorf("expected ErrProductNotFound, got %v", err)
		}
	})
}

func TestProductUsecase_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("returns merchant-scoped products", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P1"})
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P2"})
		mockProd.Create(ctx, &domain.Product{MerchantID: 2, CategoryID: 1, Name: "Other"})

		uc := NewProductUsecase(mockProd, &mockCategoryRepo{}, newMockUnitRepo())
		items, meta, err := uc.List(ctx, 1, "", pagination.Params{Offset: 0, Limit: 10})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 2 {
			t.Errorf("expected 2 items, got %d", len(items))
		}
		if meta.Total != 2 {
			t.Errorf("expected total 2, got %d", meta.Total)
		}
	})
}

func TestProductUsecase_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("partial update", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "Old"})
		uc := NewProductUsecase(mockProd, &mockCategoryRepo{}, newMockUnitRepo())

		name := "Updated"
		p, err := uc.Update(ctx, UpdateProductParams{ID: 1, Name: &name})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name != "Updated" {
			t.Errorf("expected 'Updated', got '%s'", p.Name)
		}
	})
}

func TestProductUsecase_MerchantIsolation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("merchant cannot access another merchant's product", func(t *testing.T) {
		mockProd := &mockProductRepo{}
		mockProd.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P1"})
		uc := NewProductUsecase(mockProd, &mockCategoryRepo{}, newMockUnitRepo())

		items, _, err := uc.List(ctx, 2, "", pagination.Params{Offset: 0, Limit: 10})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 0 {
			t.Errorf("expected 0 items for merchant 2, got %d", len(items))
		}
	})
}
