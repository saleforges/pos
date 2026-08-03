package usecase

import (
	"context"
	"testing"

	"github.com/saleforge/pos/services/internal/inventory/domain"
)

func TestStockUsecase_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo)

		stock, err := uc.Create(ctx, CreateStockParams{
			MerchantID:    1,
			BranchID:      1,
			ProductItemID: 1,
			Available:     100,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stock.ID == 0 {
			t.Error("expected non-zero id")
		}
		if stock.Available != 100 {
			t.Errorf("expected available 100, got %d", stock.Available)
		}
		if stock.MerchantID != 1 {
			t.Errorf("expected merchantID 1, got %d", stock.MerchantID)
		}
	})

	t.Run("missing merchant_id returns error", func(t *testing.T) {
		uc := NewStockUsecase(&mockStockRepo{})
		_, err := uc.Create(ctx, CreateStockParams{
			BranchID:      1,
			ProductItemID: 1,
			Available:     100,
		})
		if err != domain.ErrInvalidStock {
			t.Errorf("expected ErrInvalidStock, got %v", err)
		}
	})

	t.Run("missing branch_id returns error", func(t *testing.T) {
		uc := NewStockUsecase(&mockStockRepo{})
		_, err := uc.Create(ctx, CreateStockParams{
			MerchantID:    1,
			ProductItemID: 1,
			Available:     100,
		})
		if err != domain.ErrInvalidStock {
			t.Errorf("expected ErrInvalidStock, got %v", err)
		}
	})

	t.Run("missing product_item_id returns error", func(t *testing.T) {
		uc := NewStockUsecase(&mockStockRepo{})
		_, err := uc.Create(ctx, CreateStockParams{
			MerchantID: 1,
			BranchID:   1,
			Available:  100,
		})
		if err != domain.ErrInvalidStock {
			t.Errorf("expected ErrInvalidStock, got %v", err)
		}
	})

	t.Run("negative available returns error", func(t *testing.T) {
		uc := NewStockUsecase(&mockStockRepo{})
		_, err := uc.Create(ctx, CreateStockParams{
			MerchantID:    1,
			BranchID:      1,
			ProductItemID: 1,
			Available:     -1,
		})
		if err != domain.ErrNegativeAvailable {
			t.Errorf("expected ErrNegativeAvailable, got %v", err)
		}
	})
}

func TestStockUsecase_GetByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("returns stock", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo)
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 50})

		stock, err := uc.GetByID(ctx, 1, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stock.Available != 50 {
			t.Errorf("expected available 50, got %d", stock.Available)
		}
	})

	t.Run("not found returns error", func(t *testing.T) {
		uc := NewStockUsecase(&mockStockRepo{})
		_, err := uc.GetByID(ctx, 999, 1)
		if err != domain.ErrStockNotFound {
			t.Errorf("expected ErrStockNotFound, got %v", err)
		}
	})

	t.Run("different merchant returns error", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo)
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 50})

		_, err := uc.GetByID(ctx, 1, 2)
		if err != domain.ErrStockNotFound {
			t.Errorf("expected ErrStockNotFound for cross-merchant, got %v", err)
		}
	})
}

func TestStockUsecase_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("lists stocks for merchant", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo)
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 50})
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 2, ProductItemID: 1, Available: 100})
		uc.Create(ctx, CreateStockParams{MerchantID: 2, BranchID: 1, ProductItemID: 1, Available: 200})

		stocks, err := uc.List(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(stocks) != 2 {
			t.Errorf("expected 2 stocks for merchant 1, got %d", len(stocks))
		}
	})
}

func TestStockUsecase_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("updates available", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo)
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 50})

		stock, err := uc.Update(ctx, UpdateStockParams{ID: 1, MerchantID: 1, Available: 200})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stock.Available != 200 {
			t.Errorf("expected available 200, got %d", stock.Available)
		}
	})

	t.Run("cross-merchant update blocked", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo)
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 50})

		_, err := uc.Update(ctx, UpdateStockParams{ID: 1, MerchantID: 2, Available: 200})
		if err != domain.ErrStockNotFound {
			t.Errorf("expected ErrStockNotFound for cross-merchant update, got %v", err)
		}
	})

	t.Run("negative available rejected", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo)
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 50})

		_, err := uc.Update(ctx, UpdateStockParams{ID: 1, MerchantID: 1, Available: -5})
		if err != domain.ErrNegativeAvailable {
			t.Errorf("expected ErrNegativeAvailable, got %v", err)
		}
	})
}
