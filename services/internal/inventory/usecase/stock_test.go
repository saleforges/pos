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
		uc := NewStockUsecase(repo, repo)

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
		uc := NewStockUsecase(&mockStockRepo{}, &mockStockRepo{})
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
		uc := NewStockUsecase(&mockStockRepo{}, &mockStockRepo{})
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
		uc := NewStockUsecase(&mockStockRepo{}, &mockStockRepo{})
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
		uc := NewStockUsecase(&mockStockRepo{}, &mockStockRepo{})
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
		uc := NewStockUsecase(repo, repo)
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
		uc := NewStockUsecase(&mockStockRepo{}, &mockStockRepo{})
		_, err := uc.GetByID(ctx, 999, 1)
		if err != domain.ErrStockNotFound {
			t.Errorf("expected ErrStockNotFound, got %v", err)
		}
	})

	t.Run("different merchant returns error", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo, repo)
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
		uc := NewStockUsecase(repo, repo)
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
		uc := NewStockUsecase(repo, repo)
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
		uc := NewStockUsecase(repo, repo)
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 50})

		_, err := uc.Update(ctx, UpdateStockParams{ID: 1, MerchantID: 2, Available: 200})
		if err != domain.ErrStockNotFound {
			t.Errorf("expected ErrStockNotFound for cross-merchant update, got %v", err)
		}
	})

	t.Run("negative available rejected", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo, repo)
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 50})

		_, err := uc.Update(ctx, UpdateStockParams{ID: 1, MerchantID: 1, Available: -5})
		if err != domain.ErrNegativeAvailable {
			t.Errorf("expected ErrNegativeAvailable, got %v", err)
		}
	})
}

func TestStockUsecase_Deduct(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	newUCWithStock := func() (StockUsecase, *mockStockRepo) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo, repo)
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 100})
		return uc, repo
	}

	t.Run("deducts stock", func(t *testing.T) {
		uc, repo := newUCWithStock()
		err := uc.Deduct(ctx, AdjustStockParams{
			MerchantID: 1, BranchID: 1, ReferenceType: "order", ReferenceID: 42,
			Items: []AdjustStockItem{{ProductItemID: 1, Quantity: 30}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		stock, _ := repo.GetByID(ctx, 1, 1)
		if stock.Available != 70 {
			t.Errorf("expected available 70, got %d", stock.Available)
		}
	})

	t.Run("insufficient stock rejected", func(t *testing.T) {
		uc, _ := newUCWithStock()
		err := uc.Deduct(ctx, AdjustStockParams{
			MerchantID: 1, BranchID: 1, ReferenceType: "order", ReferenceID: 42,
			Items: []AdjustStockItem{{ProductItemID: 1, Quantity: 101}},
		})
		if err != domain.ErrInsufficientStock {
			t.Errorf("expected ErrInsufficientStock, got %v", err)
		}
	})

	t.Run("missing stock row returns not found", func(t *testing.T) {
		uc, _ := newUCWithStock()
		err := uc.Deduct(ctx, AdjustStockParams{
			MerchantID: 1, BranchID: 1, ReferenceType: "order", ReferenceID: 42,
			Items: []AdjustStockItem{{ProductItemID: 999, Quantity: 1}},
		})
		if err != domain.ErrStockNotFound {
			t.Errorf("expected ErrStockNotFound, got %v", err)
		}
	})

	t.Run("empty items rejected", func(t *testing.T) {
		uc, _ := newUCWithStock()
		err := uc.Deduct(ctx, AdjustStockParams{MerchantID: 1, BranchID: 1, Items: []AdjustStockItem{}})
		if err != domain.ErrInvalidStock {
			t.Errorf("expected ErrInvalidStock, got %v", err)
		}
	})
}

func TestStockUsecase_Restore(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("restores stock", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo, repo)
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 50})
		uc.Deduct(ctx, AdjustStockParams{
			MerchantID: 1, BranchID: 1, ReferenceType: "order", ReferenceID: 42,
			Items: []AdjustStockItem{{ProductItemID: 1, Quantity: 20}},
		})

		err := uc.Restore(ctx, AdjustStockParams{
			MerchantID: 1, BranchID: 1, ReferenceType: "order", ReferenceID: 42,
			Items: []AdjustStockItem{{ProductItemID: 1, Quantity: 20}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		stock, _ := repo.GetByID(ctx, 1, 1)
		if stock.Available != 50 {
			t.Errorf("expected available back to 50, got %d", stock.Available)
		}
	})
}
