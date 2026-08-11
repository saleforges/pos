package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/saleforge/pos/services/internal/inventory/domain"
)

type mockMovementRepo struct {
	movements []domain.StockMovement
}

func (m *mockMovementRepo) Create(_ context.Context, movement *domain.StockMovement) error {
	m.movements = append(m.movements, *movement)
	return nil
}

func (m *mockMovementRepo) ListByProductItem(_ context.Context, productItemID int64, merchantID int64) ([]domain.StockMovement, error) {
	var result []domain.StockMovement
	for _, mv := range m.movements {
		if mv.ProductItemID == productItemID && mv.MerchantID == merchantID {
			result = append(result, mv)
		}
	}
	return result, nil
}

func (m *mockMovementRepo) List(_ context.Context, merchantID, branchID int64, productItemID *int64, from, to *time.Time) ([]domain.StockMovement, error) {
	result := []domain.StockMovement{}
	for _, mv := range m.movements {
		if mv.MerchantID != merchantID || mv.BranchID != branchID {
			continue
		}
		if productItemID != nil && mv.ProductItemID != *productItemID {
			continue
		}
		if from != nil && mv.CreatedAt.Before(*from) {
			continue
		}
		if to != nil && mv.CreatedAt.After(*to) {
			continue
		}
		result = append(result, mv)
	}
	return result, nil
}

func TestStockMovementUsecase_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	newRepo := func() *mockMovementRepo {
		return &mockMovementRepo{movements: []domain.StockMovement{
			{ID: 1, MerchantID: 1, BranchID: 1, ProductItemID: 10, Type: domain.MovementTypeStockOut, Quantity: 2, ReferenceType: "order", CreatedAt: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)},
			{ID: 2, MerchantID: 1, BranchID: 1, ProductItemID: 20, Type: domain.MovementTypeStockIn, Quantity: 5, ReferenceType: "production", CreatedAt: time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)},
			{ID: 3, MerchantID: 1, BranchID: 2, ProductItemID: 10, Type: domain.MovementTypeStockOut, Quantity: 1, ReferenceType: "order", CreatedAt: time.Date(2026, 1, 3, 10, 0, 0, 0, time.UTC)},
		}}
	}

	t.Run("filters by branch", func(t *testing.T) {
		uc := NewStockMovementUsecase(newRepo())
		result, err := uc.List(ctx, ListMovementsParams{MerchantID: 1, BranchID: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 movements for branch 1, got %d", len(result))
		}
	})

	t.Run("filters by product item", func(t *testing.T) {
		uc := NewStockMovementUsecase(newRepo())
		productItemID := int64(10)
		result, err := uc.List(ctx, ListMovementsParams{MerchantID: 1, BranchID: 1, ProductItemID: &productItemID})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 || result[0].ID != 1 {
			t.Errorf("expected only movement 1, got %+v", result)
		}
	})

	t.Run("filters by date range", func(t *testing.T) {
		uc := NewStockMovementUsecase(newRepo())
		from := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
		result, err := uc.List(ctx, ListMovementsParams{MerchantID: 1, BranchID: 1, From: &from})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 || result[0].ID != 2 {
			t.Errorf("expected only movement 2, got %+v", result)
		}
	})

	t.Run("missing branch rejected", func(t *testing.T) {
		uc := NewStockMovementUsecase(newRepo())
		_, err := uc.List(ctx, ListMovementsParams{MerchantID: 1})
		if err != domain.ErrInvalidStock {
			t.Errorf("expected ErrInvalidStock, got %v", err)
		}
	})
}
