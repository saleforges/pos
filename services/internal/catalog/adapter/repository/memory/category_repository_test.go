package memory

import (
	"context"
	"testing"
	"time"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

func TestCategoryRepository(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	now := time.Now().UTC()

	t.Run("create and get by id", func(t *testing.T) {
		repo := NewCategoryRepository()
		c := &domain.Category{MerchantID: 1, Name: "Snack", CreatedAt: now, UpdatedAt: now}
		if err := repo.Create(ctx, c); err != nil {
			t.Fatalf("Create: %v", err)
		}
		if c.ID == 0 {
			t.Error("expected non-zero id")
		}
		got, err := repo.GetByID(ctx, c.ID, 1)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Name != "Snack" {
			t.Errorf("expected 'Snack', got '%s'", got.Name)
		}
	})

	t.Run("get by id different merchant returns not found", func(t *testing.T) {
		repo := NewCategoryRepository()
		c := &domain.Category{MerchantID: 1, Name: "Snack", CreatedAt: now, UpdatedAt: now}
		repo.Create(ctx, c)
		_, err := repo.GetByID(ctx, c.ID, 2)
		if err != domain.ErrCategoryNotFound {
			t.Errorf("expected ErrCategoryNotFound for cross-merchant, got %v", err)
		}
	})

	t.Run("list by merchant", func(t *testing.T) {
		repo := NewCategoryRepository()
		repo.Create(ctx, &domain.Category{MerchantID: 1, Name: "Food", CreatedAt: now, UpdatedAt: now})
		repo.Create(ctx, &domain.Category{MerchantID: 1, Name: "Drink", CreatedAt: now, UpdatedAt: now})
		repo.Create(ctx, &domain.Category{MerchantID: 2, Name: "Other", CreatedAt: now, UpdatedAt: now})

		items, err := repo.ListByMerchant(ctx, 1)
		if err != nil {
			t.Fatalf("ListByMerchant: %v", err)
		}
		if len(items) != 2 {
			t.Errorf("expected 2 categories, got %d", len(items))
		}
	})

	t.Run("update and delete", func(t *testing.T) {
		repo := NewCategoryRepository()
		repo.Create(ctx, &domain.Category{MerchantID: 1, Name: "Old", CreatedAt: now, UpdatedAt: now})
		c, _ := repo.GetByID(ctx, 1, 1)
		c.Name = "Updated"
		if err := repo.Update(ctx, c); err != nil {
			t.Fatalf("Update: %v", err)
		}
		if err := repo.Delete(ctx, 1, 1); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		_, err := repo.GetByID(ctx, 1, 1)
		if err != domain.ErrCategoryNotFound {
			t.Errorf("expected ErrCategoryNotFound after delete, got %v", err)
		}
	})

	t.Run("cross-merchant update blocked", func(t *testing.T) {
		repo := NewCategoryRepository()
		repo.Create(ctx, &domain.Category{MerchantID: 1, Name: "Cat1", CreatedAt: now, UpdatedAt: now})
		c := &domain.Category{ID: 1, MerchantID: 2, Name: "Hacked", CreatedAt: now, UpdatedAt: now}
		err := repo.Update(ctx, c)
		if err != domain.ErrCategoryNotFound {
			t.Errorf("expected ErrCategoryNotFound for cross-merchant update, got %v", err)
		}
	})
}
