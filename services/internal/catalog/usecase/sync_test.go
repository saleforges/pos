package usecase

import (
	"context"
	"testing"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type mockBarcodeRepo struct{}

func (m *mockBarcodeRepo) Create(_ context.Context, _ *domain.ProductItemBarcode) error { return nil }
func (m *mockBarcodeRepo) GetByBarcode(_ context.Context, _ string) (*domain.ProductItemBarcode, error) {
	return nil, domain.ErrProductItemNotFound
}
func (m *mockBarcodeRepo) ListByProductItem(_ context.Context, _ int64) ([]domain.ProductItemBarcode, error) {
	return nil, nil
}
func (m *mockBarcodeRepo) ListByMerchant(_ context.Context, _ int64) ([]domain.ProductItemBarcode, error) {
	return nil, nil
}
func (m *mockBarcodeRepo) Delete(_ context.Context, _ int64) error { return nil }

func TestSyncUsecase_Sync_BranchPrice(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	newUC := func() (SyncUsecase, *mockProductItemRepo) {
		itemRepo := &mockProductItemRepo{
			items: map[int64]*domain.ProductItem{
				1: {ID: 1, ProductID: 1, MerchantID: 1, Name: "Kopi Susu", Price: domain.Price{Amount: 15000, Currency: "IDR"}, Status: domain.ProductItemStatusActive},
			},
		}
		uc := NewSyncUsecase(&mockProductRepo{}, itemRepo, &mockCategoryRepo{}, newMockUnitRepo(), &mockBarcodeRepo{})
		return uc, itemRepo
	}

	t.Run("no branch given returns base price", func(t *testing.T) {
		uc, _ := newUC()
		result, err := uc.Sync(ctx, SyncParams{MerchantID: 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Items) != 1 || result.Items[0].Price.Amount != 15000 {
			t.Fatalf("expected base price 15000, got %+v", result.Items)
		}
	})

	t.Run("branch with no override still returns base price", func(t *testing.T) {
		uc, _ := newUC()
		result, err := uc.Sync(ctx, SyncParams{MerchantID: 1, BranchID: 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Items) != 1 || result.Items[0].Price.Amount != 15000 {
			t.Fatalf("expected base price 15000, got %+v", result.Items)
		}
	})

	t.Run("branch with override returns override price", func(t *testing.T) {
		uc, itemRepo := newUC()
		if err := itemRepo.SetBranchPrice(ctx, 1, 2, 18000, "IDR"); err != nil {
			t.Fatalf("unexpected error setting branch price: %v", err)
		}

		result, err := uc.Sync(ctx, SyncParams{MerchantID: 1, BranchID: 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Items) != 1 || result.Items[0].Price.Amount != 18000 {
			t.Fatalf("expected branch override price 18000, got %+v", result.Items)
		}

		other, err := uc.Sync(ctx, SyncParams{MerchantID: 1, BranchID: 3})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(other.Items) != 1 || other.Items[0].Price.Amount != 15000 {
			t.Fatalf("expected other branch to keep base price 15000, got %+v", other.Items)
		}
	})
}
