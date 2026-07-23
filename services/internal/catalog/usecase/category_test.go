package usecase

import (
	"context"
	"testing"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

func TestCategoryUsecase_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uc := NewCategoryUsecase(&mockCategoryRepo{})
		c, err := uc.Create(ctx, CreateCategoryParams{MerchantID: 1, Name: "Snack"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ID == 0 {
			t.Error("expected non-zero id")
		}
		if c.Name != "Snack" {
			t.Errorf("expected 'Snack', got '%s'", c.Name)
		}
		if c.MerchantID != 1 {
			t.Errorf("expected merchant 1, got %d", c.MerchantID)
		}
	})

	t.Run("empty name returns error", func(t *testing.T) {
		uc := NewCategoryUsecase(&mockCategoryRepo{})
		_, err := uc.Create(ctx, CreateCategoryParams{MerchantID: 1})
		if err != domain.ErrInvalidCategory {
			t.Errorf("expected ErrInvalidCategory, got %v", err)
		}
	})

	t.Run("invalid parent returns error", func(t *testing.T) {
		parentID := int64(999)
		uc := NewCategoryUsecase(&mockCategoryRepo{})
		_, err := uc.Create(ctx, CreateCategoryParams{MerchantID: 1, Name: "Sub", ParentID: &parentID})
		if err != domain.ErrCategoryNotFound {
			t.Errorf("expected ErrCategoryNotFound, got %v", err)
		}
	})

	t.Run("merchant cannot use another merchant's category as parent", func(t *testing.T) {
		mockCat := &mockCategoryRepo{categories: map[int64]*domain.Category{
			1: {ID: 1, MerchantID: 1, Name: "Food"},
		}}
		uc := NewCategoryUsecase(mockCat)
		parentID := int64(1)
		_, err := uc.Create(ctx, CreateCategoryParams{MerchantID: 2, Name: "Sub", ParentID: &parentID})
		if err != domain.ErrCategoryNotFound {
			t.Errorf("expected ErrCategoryNotFound for cross-merchant parent, got %v", err)
		}
	})
}

func TestCategoryUsecase_ListByMerchant(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("merchant scoped", func(t *testing.T) {
		mockCat := &mockCategoryRepo{categories: map[int64]*domain.Category{
			1: {ID: 1, MerchantID: 1, Name: "Food"},
			2: {ID: 2, MerchantID: 1, Name: "Drink"},
			3: {ID: 3, MerchantID: 2, Name: "Other"},
		}}
		uc := NewCategoryUsecase(mockCat)
		items, err := uc.ListByMerchant(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(items) != 2 {
			t.Errorf("expected 2 categories, got %d", len(items))
		}
	})
}

func TestCategoryUsecase_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("rename category", func(t *testing.T) {
		mockCat := &mockCategoryRepo{}
		mockCat.Create(ctx, &domain.Category{MerchantID: 1, Name: "Old"})
		uc := NewCategoryUsecase(mockCat)

		name := "Renamed"
		c, err := uc.Update(ctx, UpdateCategoryParams{ID: 1, MerchantID: 1, Name: &name})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Name != "Renamed" {
			t.Errorf("expected 'Renamed', got '%s'", c.Name)
		}
	})

	t.Run("merchant cannot update another merchant's category", func(t *testing.T) {
		mockCat := &mockCategoryRepo{}
		mockCat.Create(ctx, &domain.Category{MerchantID: 1, Name: "Snack"})
		uc := NewCategoryUsecase(mockCat)

		name := "Hacked"
		_, err := uc.Update(ctx, UpdateCategoryParams{ID: 1, MerchantID: 2, Name: &name})
		if err != domain.ErrCategoryNotFound {
			t.Errorf("expected ErrCategoryNotFound for cross-merchant update, got %v", err)
		}
	})
}

func TestCategoryUsecase_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("merchant cannot delete another merchant's category", func(t *testing.T) {
		mockCat := &mockCategoryRepo{}
		mockCat.Create(ctx, &domain.Category{MerchantID: 1, Name: "Snack"})
		uc := NewCategoryUsecase(mockCat)

		err := uc.Delete(ctx, 1, 2)
		if err != domain.ErrCategoryNotFound {
			t.Errorf("expected ErrCategoryNotFound for cross-merchant delete, got %v", err)
		}
	})

	t.Run("own category delete succeeds", func(t *testing.T) {
		mockCat := &mockCategoryRepo{}
		mockCat.Create(ctx, &domain.Category{MerchantID: 1, Name: "Snack"})
		uc := NewCategoryUsecase(mockCat)

		err := uc.Delete(ctx, 1, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
