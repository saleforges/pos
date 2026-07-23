package memory

import (
	"context"
	"testing"
	"time"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

func TestProductRepository(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	now := time.Now().UTC()

	t.Run("create and get by id", func(t *testing.T) {
		repo := NewProductRepository()
		p := &domain.Product{MerchantID: 1, CategoryID: 1, Name: "Test Product", Description: "Desc", Status: domain.ProductStatusActive, CreatedAt: now, UpdatedAt: now}
		if err := repo.Create(ctx, p); err != nil {
			t.Fatalf("Create: %v", err)
		}
		if p.ID == 0 {
			t.Error("expected non-zero id")
		}
		got, err := repo.GetByID(ctx, p.ID)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Name != "Test Product" {
			t.Errorf("expected 'Test Product', got '%s'", got.Name)
		}
	})

	t.Run("get by id not found", func(t *testing.T) {
		repo := NewProductRepository()
		_, err := repo.GetByID(ctx, 999)
		if err != domain.ErrProductNotFound {
			t.Errorf("expected ErrProductNotFound, got %v", err)
		}
	})

	t.Run("list by merchant with search", func(t *testing.T) {
		repo := NewProductRepository()
		repo.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "Marning", Status: domain.ProductStatusActive, CreatedAt: now, UpdatedAt: now})
		repo.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "Aqua", Status: domain.ProductStatusActive, CreatedAt: now, UpdatedAt: now})
		repo.Create(ctx, &domain.Product{MerchantID: 2, CategoryID: 1, Name: "Other", Status: domain.ProductStatusActive, CreatedAt: now, UpdatedAt: now})

		items, total, err := repo.List(ctx, 1, "", 0, 10)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total != 2 {
			t.Errorf("expected total 2, got %d", total)
		}
		if len(items) != 2 {
			t.Errorf("expected 2 items, got %d", len(items))
		}

		items, total, err = repo.List(ctx, 1, "Marning", 0, 10)
		if err != nil {
			t.Fatalf("List with search: %v", err)
		}
		if total != 1 {
			t.Errorf("expected total 1, got %d", total)
		}
	})

	t.Run("update", func(t *testing.T) {
		repo := NewProductRepository()
		repo.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "Old", Status: domain.ProductStatusActive, CreatedAt: now, UpdatedAt: now})
		p, _ := repo.GetByID(ctx, 1)
		p.Name = "New"
		p.UpdatedAt = time.Now().UTC()

		if err := repo.Update(ctx, p); err != nil {
			t.Fatalf("Update: %v", err)
		}
		got, _ := repo.GetByID(ctx, 1)
		if got.Name != "New" {
			t.Errorf("expected 'New', got '%s'", got.Name)
		}
	})

	t.Run("delete", func(t *testing.T) {
		repo := NewProductRepository()
		repo.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "Del", Status: domain.ProductStatusActive, CreatedAt: now, UpdatedAt: now})
		if err := repo.Delete(ctx, 1); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		_, err := repo.GetByID(ctx, 1)
		if err != domain.ErrProductNotFound {
			t.Errorf("expected ErrProductNotFound after delete, got %v", err)
		}
	})

	t.Run("pagination", func(t *testing.T) {
		repo := NewProductRepository()
		for i := 0; i < 5; i++ {
			repo.Create(ctx, &domain.Product{MerchantID: 1, CategoryID: 1, Name: "P", Status: domain.ProductStatusActive, CreatedAt: now, UpdatedAt: now})
		}
		items, total, err := repo.List(ctx, 1, "", 0, 2)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if total != 5 {
			t.Errorf("expected total 5, got %d", total)
		}
		if len(items) != 2 {
			t.Errorf("expected 2 items, got %d", len(items))
		}
	})
}
