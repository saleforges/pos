package memory

import (
	"context"
	"testing"
	"time"

	"github.com/saleforge/pos/services/internal/inventory/domain"
)

func TestStockRepository(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	now := time.Now().UTC()

	t.Run("create and get by id", func(t *testing.T) {
		repo := NewStockRepository()
		stock := &domain.Stock{
			MerchantID:    1,
			BranchID:      1,
			ProductItemID: 1,
			Available:     100,
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		if err := repo.Create(ctx, stock); err != nil {
			t.Fatalf("Create: %v", err)
		}
		if stock.ID == 0 {
			t.Error("expected non-zero id")
		}
		got, err := repo.GetByID(ctx, stock.ID, 1)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Available != 100 {
			t.Errorf("expected available 100, got %d", got.Available)
		}
	})

	t.Run("get by id different merchant returns not found", func(t *testing.T) {
		repo := NewStockRepository()
		stock := &domain.Stock{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 50, CreatedAt: now, UpdatedAt: now}
		repo.Create(ctx, stock)
		_, err := repo.GetByID(ctx, stock.ID, 2)
		if err != domain.ErrStockNotFound {
			t.Errorf("expected ErrStockNotFound for cross-merchant, got %v", err)
		}
	})

	t.Run("get by id not found", func(t *testing.T) {
		repo := NewStockRepository()
		_, err := repo.GetByID(ctx, 999, 1)
		if err != domain.ErrStockNotFound {
			t.Errorf("expected ErrStockNotFound, got %v", err)
		}
	})

	t.Run("list by merchant", func(t *testing.T) {
		repo := NewStockRepository()
		repo.Create(ctx, &domain.Stock{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 100, CreatedAt: now, UpdatedAt: now})
		repo.Create(ctx, &domain.Stock{MerchantID: 1, BranchID: 2, ProductItemID: 2, Available: 200, CreatedAt: now, UpdatedAt: now})
		repo.Create(ctx, &domain.Stock{MerchantID: 2, BranchID: 1, ProductItemID: 3, Available: 300, CreatedAt: now, UpdatedAt: now})

		stocks, err := repo.List(ctx, 1)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(stocks) != 2 {
			t.Errorf("expected 2 stocks for merchant 1, got %d", len(stocks))
		}
	})

	t.Run("update", func(t *testing.T) {
		repo := NewStockRepository()
		repo.Create(ctx, &domain.Stock{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 100, CreatedAt: now, UpdatedAt: now})
		stock, _ := repo.GetByID(ctx, 1, 1)
		stock.Available = 250
		if err := repo.Update(ctx, stock); err != nil {
			t.Fatalf("Update: %v", err)
		}
		got, _ := repo.GetByID(ctx, 1, 1)
		if got.Available != 250 {
			t.Errorf("expected available 250, got %d", got.Available)
		}
	})

	t.Run("cross-merchant update blocked", func(t *testing.T) {
		repo := NewStockRepository()
		repo.Create(ctx, &domain.Stock{MerchantID: 1, BranchID: 1, ProductItemID: 1, Available: 100, CreatedAt: now, UpdatedAt: now})
		stock := &domain.Stock{ID: 1, MerchantID: 2, BranchID: 1, ProductItemID: 1, Available: 999, CreatedAt: now, UpdatedAt: now}
		err := repo.Update(ctx, stock)
		if err != domain.ErrStockNotFound {
			t.Errorf("expected ErrStockNotFound for cross-merchant update, got %v", err)
		}
	})
}
