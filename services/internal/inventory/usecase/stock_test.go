package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/saleforge/pos/services/internal/inventory/domain"
)

func TestStockUsecase_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo, repo, &mockComponentRepo{}, newMockUnitRepo())

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
		uc := NewStockUsecase(&mockStockRepo{}, &mockStockRepo{}, &mockComponentRepo{}, newMockUnitRepo())
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
		uc := NewStockUsecase(&mockStockRepo{}, &mockStockRepo{}, &mockComponentRepo{}, newMockUnitRepo())
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
		uc := NewStockUsecase(&mockStockRepo{}, &mockStockRepo{}, &mockComponentRepo{}, newMockUnitRepo())
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
		uc := NewStockUsecase(&mockStockRepo{}, &mockStockRepo{}, &mockComponentRepo{}, newMockUnitRepo())
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
		uc := NewStockUsecase(repo, repo, &mockComponentRepo{}, newMockUnitRepo())
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
		uc := NewStockUsecase(&mockStockRepo{}, &mockStockRepo{}, &mockComponentRepo{}, newMockUnitRepo())
		_, err := uc.GetByID(ctx, 999, 1)
		if err != domain.ErrStockNotFound {
			t.Errorf("expected ErrStockNotFound, got %v", err)
		}
	})

	t.Run("different merchant returns error", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo, repo, &mockComponentRepo{}, newMockUnitRepo())
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
		uc := NewStockUsecase(repo, repo, &mockComponentRepo{}, newMockUnitRepo())
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
		uc := NewStockUsecase(repo, repo, &mockComponentRepo{}, newMockUnitRepo())
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
		uc := NewStockUsecase(repo, repo, &mockComponentRepo{}, newMockUnitRepo())
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 50})

		_, err := uc.Update(ctx, UpdateStockParams{ID: 1, MerchantID: 2, Available: 200})
		if err != domain.ErrStockNotFound {
			t.Errorf("expected ErrStockNotFound for cross-merchant update, got %v", err)
		}
	})

	t.Run("negative available rejected", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo, repo, &mockComponentRepo{}, newMockUnitRepo())
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
		uc := NewStockUsecase(repo, repo, &mockComponentRepo{}, newMockUnitRepo())
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
		uc := NewStockUsecase(repo, repo, &mockComponentRepo{}, newMockUnitRepo())
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

func TestStockUsecase_ComponentConsumption(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	newUC := func() (StockUsecase, *mockStockRepo, *mockComponentRepo) {
		repo := &mockStockRepo{}
		compRepo := &mockComponentRepo{}
		uc := NewStockUsecase(repo, repo, compRepo, newMockUnitRepo())
		// stock: item 1 (produk jadi) = 10, item 2 (bahan baku) = 20
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 10})
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 2, Available: 20})
		// component: 1x item 1 = 2x item 2
		compRepo.Create(ctx, &domain.ProductComponent{
			MerchantID:    1,
			ProductItemID: 1,
			Items: []domain.ProductComponentItem{
				{ComponentProductItemID: 2, Quantity: 2, UnitID: 1},
			},
		})
		return uc, repo, compRepo
	}

	t.Run("deducts raw materials too", func(t *testing.T) {
		uc, repo, _ := newUC()
		err := uc.Deduct(ctx, AdjustStockParams{
			MerchantID: 1, BranchID: 1, ReferenceType: "order", ReferenceID: 42,
			Items: []AdjustStockItem{{ProductItemID: 1, Quantity: 3}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		product, _ := repo.GetByID(ctx, 1, 1)
		raw, _ := repo.GetByID(ctx, 2, 1)
		if product.Available != 7 {
			t.Errorf("expected product stock 7, got %d", product.Available)
		}
		if raw.Available != 14 {
			t.Errorf("expected raw material stock 14 (20-3*2), got %d", raw.Available)
		}
	})

	t.Run("insufficient raw material rejected", func(t *testing.T) {
		uc, _, _ := newUC()
		// need 10*2=20 raw material but only 20 available → exactly enough, so use 11
		err := uc.Deduct(ctx, AdjustStockParams{
			MerchantID: 1, BranchID: 1, ReferenceType: "order", ReferenceID: 42,
			Items: []AdjustStockItem{{ProductItemID: 1, Quantity: 11}},
		})
		if err != domain.ErrInsufficientStock {
			t.Errorf("expected ErrInsufficientStock, got %v", err)
		}
	})

	t.Run("no component = normal deduct", func(t *testing.T) {
		uc, repo, _ := newUC()
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 99, Available: 5})
		err := uc.Deduct(ctx, AdjustStockParams{
			MerchantID: 1, BranchID: 1, ReferenceType: "order", ReferenceID: 42,
			Items: []AdjustStockItem{{ProductItemID: 99, Quantity: 2}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		stock, _ := repo.GetByID(ctx, 3, 1)
		if stock.Available != 3 {
			t.Errorf("expected 3, got %d", stock.Available)
		}
	})

	t.Run("restore returns raw materials", func(t *testing.T) {
		uc, repo, _ := newUC()
		uc.Deduct(ctx, AdjustStockParams{
			MerchantID: 1, BranchID: 1, ReferenceType: "order", ReferenceID: 42,
			Items: []AdjustStockItem{{ProductItemID: 1, Quantity: 3}},
		})
		err := uc.Restore(ctx, AdjustStockParams{
			MerchantID: 1, BranchID: 1, ReferenceType: "order", ReferenceID: 42,
			Items: []AdjustStockItem{{ProductItemID: 1, Quantity: 3}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		product, _ := repo.GetByID(ctx, 1, 1)
		raw, _ := repo.GetByID(ctx, 2, 1)
		if product.Available != 10 {
			t.Errorf("expected product stock back to 10, got %d", product.Available)
		}
		if raw.Available != 20 {
			t.Errorf("expected raw material stock back to 20, got %d", raw.Available)
		}
	})
}

func TestStockUsecase_UnitConversion(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("component quantity normalized via unit factor", func(t *testing.T) {
		repo := &mockStockRepo{}
		compRepo := &mockComponentRepo{}
		uc := NewStockUsecase(repo, repo, compRepo, newMockUnitRepo())
		// stock: item 1 (produk jadi) = 10, item 2 (gula) = 30000 gram
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 10})
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 2, Available: 30000})
		// 1x item 1 = 0.5 kg gula; kg factor-to-base 1000 → 500 gram per item
		compRepo.Create(ctx, &domain.ProductComponent{
			MerchantID:    1,
			ProductItemID: 1,
			Items: []domain.ProductComponentItem{
				{ComponentProductItemID: 2, Quantity: 0.5, UnitID: 3},
			},
		})

		err := uc.Deduct(ctx, AdjustStockParams{
			MerchantID: 1, BranchID: 1, ReferenceType: "order", ReferenceID: 42,
			Items: []AdjustStockItem{{ProductItemID: 1, Quantity: 2}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		raw, _ := repo.GetByID(ctx, 2, 1)
		// 2 x 0.5 kg x 1000 = 1000 gram deducted
		if raw.Available != 29000 {
			t.Errorf("expected raw stock 29000 (30000-1000), got %d", raw.Available)
		}
	})

	t.Run("shared raw material merged after normalization", func(t *testing.T) {
		repo := &mockStockRepo{}
		compRepo := &mockComponentRepo{}
		uc := NewStockUsecase(repo, repo, compRepo, newMockUnitRepo())
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 10})
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 2, Available: 30000})
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 3, Available: 10})
		// both items 1 and 3 consume item 2: 0.5 kg each
		compRepo.Create(ctx, &domain.ProductComponent{
			MerchantID: 1, ProductItemID: 1,
			Items: []domain.ProductComponentItem{
				{ComponentProductItemID: 2, Quantity: 0.5, UnitID: 3},
			},
		})
		compRepo.Create(ctx, &domain.ProductComponent{
			MerchantID: 1, ProductItemID: 3,
			Items: []domain.ProductComponentItem{
				{ComponentProductItemID: 2, Quantity: 0.5, UnitID: 3},
			},
		})

		err := uc.Deduct(ctx, AdjustStockParams{
			MerchantID: 1, BranchID: 1, ReferenceType: "order", ReferenceID: 42,
			Items: []AdjustStockItem{
				{ProductItemID: 1, Quantity: 1},
				{ProductItemID: 3, Quantity: 1},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		raw, _ := repo.GetByID(ctx, 2, 1)
		// 500 + 500 = 1000 gram, single merged deduction
		if raw.Available != 29000 {
			t.Errorf("expected raw stock 29000 (30000-1000), got %d", raw.Available)
		}
	})

	t.Run("gram unit keeps quantity as-is", func(t *testing.T) {
		repo := &mockStockRepo{}
		compRepo := &mockComponentRepo{}
		uc := NewStockUsecase(repo, repo, compRepo, newMockUnitRepo())
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 10})
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 2, Available: 30000})
		compRepo.Create(ctx, &domain.ProductComponent{
			MerchantID: 1, ProductItemID: 1,
			Items: []domain.ProductComponentItem{
				{ComponentProductItemID: 2, Quantity: 500, UnitID: 4},
			},
		})

		err := uc.Deduct(ctx, AdjustStockParams{
			MerchantID: 1, BranchID: 1, ReferenceType: "order", ReferenceID: 42,
			Items: []AdjustStockItem{{ProductItemID: 1, Quantity: 2}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		raw, _ := repo.GetByID(ctx, 2, 1)
		// 2 x 500 gram x 1 = 1000 gram
		if raw.Available != 29000 {
			t.Errorf("expected raw stock 29000, got %d", raw.Available)
		}
	})

	t.Run("unknown unit falls back to factor 1", func(t *testing.T) {
		repo := &mockStockRepo{}
		compRepo := &mockComponentRepo{}
		uc := NewStockUsecase(repo, repo, compRepo, newMockUnitRepo())
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 10})
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 2, Available: 30000})
		compRepo.Create(ctx, &domain.ProductComponent{
			MerchantID: 1, ProductItemID: 1,
			Items: []domain.ProductComponentItem{
				{ComponentProductItemID: 2, Quantity: 500, UnitID: 999},
			},
		})

		err := uc.Deduct(ctx, AdjustStockParams{
			MerchantID: 1, BranchID: 1, ReferenceType: "order", ReferenceID: 42,
			Items: []AdjustStockItem{{ProductItemID: 1, Quantity: 2}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		raw, _ := repo.GetByID(ctx, 2, 1)
		// unknown unit → factor 1 → 2 x 500 = 1000
		if raw.Available != 29000 {
			t.Errorf("expected raw stock 29000, got %d", raw.Available)
		}
	})
}

func TestStockUsecase_Sync(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("returns changes since lastSync with token", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo, repo, &mockComponentRepo{}, newMockUnitRepo())
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 10})

		since := time.Now().UTC().Add(-time.Hour)
		result, err := uc.Sync(ctx, 1, &since)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Stocks) != 1 {
			t.Errorf("expected 1 stock, got %d", len(result.Stocks))
		}
		if result.SyncToken == "" {
			t.Error("expected non-empty sync token")
		}
	})

	t.Run("no lastSync returns all", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo, repo, &mockComponentRepo{}, newMockUnitRepo())
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 10})
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 2, Available: 20})

		result, err := uc.Sync(ctx, 1, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Stocks) != 2 {
			t.Errorf("expected 2 stocks, got %d", len(result.Stocks))
		}
	})
}
