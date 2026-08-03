package memory

import (
	"context"
	"testing"
	"time"

	"github.com/saleforge/pos/services/internal/inventory/domain"
)

func TestProductComponentRepository(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	now := time.Now().UTC()

	t.Run("create and get by product item", func(t *testing.T) {
		repo := NewProductComponentRepository()
		component := &domain.ProductComponent{
			MerchantID:    1,
			ProductItemID: 1,
			CreatedAt:     now,
			UpdatedAt:     now,
			Items: []domain.ProductComponentItem{
				{ComponentProductItemID: 2, Quantity: 2, UnitID: 1, CreatedAt: now},
				{ComponentProductItemID: 3, Quantity: 1, UnitID: 2, CreatedAt: now},
			},
		}
		if err := repo.Create(ctx, component); err != nil {
			t.Fatalf("Create: %v", err)
		}
		if component.ID == 0 {
			t.Error("expected non-zero id")
		}
		got, err := repo.GetByProductItem(ctx, 1, 1)
		if err != nil {
			t.Fatalf("GetByProductItem: %v", err)
		}
		if got.ProductItemID != 1 {
			t.Errorf("expected productItemID 1, got %d", got.ProductItemID)
		}
		if len(got.Items) != 2 {
			t.Errorf("expected 2 items, got %d", len(got.Items))
		}
	})

	t.Run("duplicate product item for same merchant rejected", func(t *testing.T) {
		repo := NewProductComponentRepository()
		repo.Create(ctx, &domain.ProductComponent{
			MerchantID: 1, ProductItemID: 1, CreatedAt: now, UpdatedAt: now,
			Items: []domain.ProductComponentItem{{ComponentProductItemID: 2, Quantity: 1, UnitID: 1, CreatedAt: now}},
		})
		err := repo.Create(ctx, &domain.ProductComponent{
			MerchantID: 1, ProductItemID: 1, CreatedAt: now, UpdatedAt: now,
			Items: []domain.ProductComponentItem{{ComponentProductItemID: 3, Quantity: 2, UnitID: 1, CreatedAt: now}},
		})
		if err != domain.ErrComponentAlreadyExists {
			t.Errorf("expected ErrComponentAlreadyExists, got %v", err)
		}
	})

	t.Run("different merchant can have component for same product item", func(t *testing.T) {
		repo := NewProductComponentRepository()
		repo.Create(ctx, &domain.ProductComponent{
			MerchantID: 1, ProductItemID: 1, CreatedAt: now, UpdatedAt: now,
			Items: []domain.ProductComponentItem{{ComponentProductItemID: 2, Quantity: 1, UnitID: 1, CreatedAt: now}},
		})
		err := repo.Create(ctx, &domain.ProductComponent{
			MerchantID: 2, ProductItemID: 1, CreatedAt: now, UpdatedAt: now,
			Items: []domain.ProductComponentItem{{ComponentProductItemID: 3, Quantity: 2, UnitID: 1, CreatedAt: now}},
		})
		if err != nil {
			t.Fatalf("expected success for different merchant, got %v", err)
		}
	})

	t.Run("list by merchant", func(t *testing.T) {
		repo := NewProductComponentRepository()
		repo.Create(ctx, &domain.ProductComponent{
			MerchantID: 1, ProductItemID: 1, CreatedAt: now, UpdatedAt: now,
			Items: []domain.ProductComponentItem{{ComponentProductItemID: 2, Quantity: 1, UnitID: 1, CreatedAt: now}},
		})
		repo.Create(ctx, &domain.ProductComponent{
			MerchantID: 2, ProductItemID: 2, CreatedAt: now, UpdatedAt: now,
			Items: []domain.ProductComponentItem{{ComponentProductItemID: 3, Quantity: 2, UnitID: 1, CreatedAt: now}},
		})

		components, err := repo.List(ctx, 1)
		if err != nil {
			t.Fatalf("List: %v", err)
		}
		if len(components) != 1 {
			t.Errorf("expected 1 component for merchant 1, got %d", len(components))
		}
	})

	t.Run("update", func(t *testing.T) {
		repo := NewProductComponentRepository()
		repo.Create(ctx, &domain.ProductComponent{
			MerchantID: 1, ProductItemID: 1, CreatedAt: now, UpdatedAt: now,
			Items: []domain.ProductComponentItem{{ComponentProductItemID: 2, Quantity: 1, UnitID: 1, CreatedAt: now}},
		})
		component, _ := repo.GetByProductItem(ctx, 1, 1)
		component.Items = []domain.ProductComponentItem{
			{ComponentProductItemID: 3, Quantity: 3, UnitID: 2, CreatedAt: now},
		}
		if err := repo.Update(ctx, component); err != nil {
			t.Fatalf("Update: %v", err)
		}
		got, _ := repo.GetByProductItem(ctx, 1, 1)
		if len(got.Items) != 1 {
			t.Errorf("expected 1 item, got %d", len(got.Items))
		}
		if got.Items[0].Quantity != 3 {
			t.Errorf("expected quantity 3, got %f", got.Items[0].Quantity)
		}
	})

	t.Run("cross-merchant update blocked", func(t *testing.T) {
		repo := NewProductComponentRepository()
		repo.Create(ctx, &domain.ProductComponent{
			MerchantID: 1, ProductItemID: 1, CreatedAt: now, UpdatedAt: now,
			Items: []domain.ProductComponentItem{{ComponentProductItemID: 2, Quantity: 1, UnitID: 1, CreatedAt: now}},
		})
		err := repo.Update(ctx, &domain.ProductComponent{
			ID: 1, MerchantID: 2, ProductItemID: 1, CreatedAt: now, UpdatedAt: now,
		})
		if err != domain.ErrProductComponentNotFound {
			t.Errorf("expected ErrProductComponentNotFound for cross-merchant update, got %v", err)
		}
	})

	t.Run("delete", func(t *testing.T) {
		repo := NewProductComponentRepository()
		repo.Create(ctx, &domain.ProductComponent{
			MerchantID: 1, ProductItemID: 1, CreatedAt: now, UpdatedAt: now,
			Items: []domain.ProductComponentItem{{ComponentProductItemID: 2, Quantity: 1, UnitID: 1, CreatedAt: now}},
		})
		if err := repo.Delete(ctx, 1, 1); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		_, err := repo.GetByProductItem(ctx, 1, 1)
		if err != domain.ErrProductComponentNotFound {
			t.Errorf("expected ErrProductComponentNotFound after delete, got %v", err)
		}
	})

	t.Run("cross-merchant delete blocked", func(t *testing.T) {
		repo := NewProductComponentRepository()
		repo.Create(ctx, &domain.ProductComponent{
			MerchantID: 1, ProductItemID: 1, CreatedAt: now, UpdatedAt: now,
			Items: []domain.ProductComponentItem{{ComponentProductItemID: 2, Quantity: 1, UnitID: 1, CreatedAt: now}},
		})
		err := repo.Delete(ctx, 1, 2)
		if err != domain.ErrProductComponentNotFound {
			t.Errorf("expected ErrProductComponentNotFound for cross-merchant delete, got %v", err)
		}
	})
}
