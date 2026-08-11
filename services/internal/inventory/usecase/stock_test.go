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

	t.Run("missing stock row is skipped, not an error", func(t *testing.T) {
		uc, _ := newUCWithStock()
		err := uc.Deduct(ctx, AdjustStockParams{
			MerchantID: 1, BranchID: 1, ReferenceType: "order", ReferenceID: 42,
			Items: []AdjustStockItem{{ProductItemID: 999, Quantity: 1}},
		})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
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
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 10})
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 2, Available: 20})
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

	t.Run("derived item with no stock row of its own still deducts its component", func(t *testing.T) {
		repo := &mockStockRepo{}
		compRepo := &mockComponentRepo{}
		uc := NewStockUsecase(repo, repo, compRepo, newMockUnitRepo())
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 2, Available: 10})
		compRepo.Create(ctx, &domain.ProductComponent{
			MerchantID:    1,
			ProductItemID: 1,
			Items: []domain.ProductComponentItem{
				{ComponentProductItemID: 2, Quantity: 1, UnitID: 1},
			},
		})

		err := uc.Deduct(ctx, AdjustStockParams{
			MerchantID: 1, BranchID: 1, ReferenceType: "order", ReferenceID: 42,
			Items: []AdjustStockItem{{ProductItemID: 1, Quantity: 2}},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		raw, _ := repo.GetByID(ctx, 1, 1)
		if raw.Available != 8 {
			t.Errorf("expected component stock 8 (10-2), got %d", raw.Available)
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
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 10})
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 2, Available: 30000})
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
		if raw.Available != 29000 {
			t.Errorf("expected raw stock 29000, got %d", raw.Available)
		}
	})
}

func TestStockUsecase_Opname(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("creates stock row when none exists", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo, repo, &mockComponentRepo{}, newMockUnitRepo())

		stock, err := uc.Opname(ctx, OpnameParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, ActualQuantity: 40, Reason: "opname"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stock.Available != 40 {
			t.Errorf("expected available 40, got %d", stock.Available)
		}
	})

	t.Run("increases stock when actual is higher", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo, repo, &mockComponentRepo{}, newMockUnitRepo())
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 50})

		stock, err := uc.Opname(ctx, OpnameParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, ActualQuantity: 55, Reason: "opname"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stock.Available != 55 {
			t.Errorf("expected available 55, got %d", stock.Available)
		}
	})

	t.Run("decreases stock when actual is lower", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo, repo, &mockComponentRepo{}, newMockUnitRepo())
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 100})

		stock, err := uc.Opname(ctx, OpnameParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, ActualQuantity: 98, Reason: "internal_use"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stock.Available != 98 {
			t.Errorf("expected available 98, got %d", stock.Available)
		}
	})

	t.Run("no-op when actual matches current", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo, repo, &mockComponentRepo{}, newMockUnitRepo())
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 100})

		stock, err := uc.Opname(ctx, OpnameParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, ActualQuantity: 100})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stock.Available != 100 {
			t.Errorf("expected available 100, got %d", stock.Available)
		}
	})

	t.Run("negative actual quantity rejected", func(t *testing.T) {
		uc := NewStockUsecase(&mockStockRepo{}, &mockStockRepo{}, &mockComponentRepo{}, newMockUnitRepo())
		_, err := uc.Opname(ctx, OpnameParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, ActualQuantity: -1})
		if err != domain.ErrInvalidStock {
			t.Errorf("expected ErrInvalidStock, got %v", err)
		}
	})
}

func TestStockUsecase_Produce(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	newUC := func() (StockUsecase, *mockStockRepo, *mockComponentRepo) {
		repo := &mockStockRepo{}
		compRepo := &mockComponentRepo{}
		uc := NewStockUsecase(repo, repo, compRepo, newMockUnitRepo())
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 2, Available: 100})
		compRepo.Create(ctx, &domain.ProductComponent{
			MerchantID:    1,
			ProductItemID: 1,
			Items: []domain.ProductComponentItem{
				{ComponentProductItemID: 2, Quantity: 1, UnitID: 1},
			},
		})
		return uc, repo, compRepo
	}

	t.Run("produces stock from raw material and clears the recipe", func(t *testing.T) {
		uc, repo, compRepo := newUC()

		produced, err := uc.Produce(ctx, ProduceParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Quantity: 100})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if produced.Available != 100 {
			t.Errorf("expected produced stock 100, got %d", produced.Available)
		}

		raw, _ := repo.GetByID(ctx, 1, 1)
		if raw.Available != 0 {
			t.Errorf("expected raw material stock 0 (100-100), got %d", raw.Available)
		}

		if _, err := compRepo.GetByProductItem(ctx, 1, 1); err != domain.ErrProductComponentNotFound {
			t.Errorf("expected recipe to be deleted after produce, got %v", err)
		}
	})

	t.Run("insufficient raw material rejects produce", func(t *testing.T) {
		uc, repo, compRepo := newUC()
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 5})

		_, err := uc.Produce(ctx, ProduceParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Quantity: 150})
		if err != domain.ErrInsufficientStock {
			t.Fatalf("expected ErrInsufficientStock, got %v", err)
		}
		raw, _ := repo.GetByID(ctx, 1, 1)
		if raw.Available != 100 {
			t.Errorf("expected raw material stock untouched 100, got %d", raw.Available)
		}
		if _, err := compRepo.GetByProductItem(ctx, 1, 1); err != nil {
			t.Errorf("expected recipe kept after failed produce, got %v", err)
		}
	})

	t.Run("increments existing stock row when already present", func(t *testing.T) {
		uc, repo, _ := newUC()
		uc.Create(ctx, CreateStockParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 20})

		produced, err := uc.Produce(ctx, ProduceParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Quantity: 30})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if produced.Available != 50 {
			t.Errorf("expected produced stock 50 (20+30), got %d", produced.Available)
		}
		raw, _ := repo.GetByID(ctx, 1, 1)
		if raw.Available != 70 {
			t.Errorf("expected raw material stock 70 (100-30), got %d", raw.Available)
		}
	})

	t.Run("missing recipe returns error", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo, repo, &mockComponentRepo{}, newMockUnitRepo())

		_, err := uc.Produce(ctx, ProduceParams{MerchantID: 1, BranchID: 1, ProductItemID: 9, Quantity: 5})
		if err != domain.ErrProductComponentNotFound {
			t.Errorf("expected ErrProductComponentNotFound, got %v", err)
		}
	})

	t.Run("invalid params rejected", func(t *testing.T) {
		repo := &mockStockRepo{}
		uc := NewStockUsecase(repo, repo, &mockComponentRepo{}, newMockUnitRepo())
		if _, err := uc.Produce(ctx, ProduceParams{MerchantID: 1, BranchID: 1, ProductItemID: 1, Quantity: 0}); err != domain.ErrInvalidStock {
			t.Errorf("expected ErrInvalidStock for zero quantity, got %v", err)
		}
		if _, err := uc.Produce(ctx, ProduceParams{MerchantID: 1, BranchID: 0, ProductItemID: 1, Quantity: 5}); err != domain.ErrInvalidStock {
			t.Errorf("expected ErrInvalidStock for missing branch, got %v", err)
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
		result, err := uc.Sync(ctx, 1, 0, &since)
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

		result, err := uc.Sync(ctx, 1, 0, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Stocks) != 2 {
			t.Errorf("expected 2 stocks, got %d", len(result.Stocks))
		}
	})
}
