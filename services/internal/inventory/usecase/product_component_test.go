package usecase

import (
	"context"
	"testing"

	"github.com/saleforge/pos/services/internal/inventory/domain"
)

func TestProductComponentUsecase_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := &mockComponentRepo{}
		uc := NewProductComponentUsecase(repo)

		component, err := uc.Create(ctx, CreateProductComponentParams{
			MerchantID:    1,
			ProductItemID: 1,
			Items: []CreateProductComponentItemParams{
				{ComponentProductItemID: 2, Quantity: 2, UnitID: 1},
				{ComponentProductItemID: 3, Quantity: 1, UnitID: 1},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if component.ID == 0 {
			t.Error("expected non-zero id")
		}
		if len(component.Items) != 2 {
			t.Errorf("expected 2 items, got %d", len(component.Items))
		}
	})

	t.Run("missing merchant_id returns error", func(t *testing.T) {
		uc := NewProductComponentUsecase(&mockComponentRepo{})
		_, err := uc.Create(ctx, CreateProductComponentParams{
			ProductItemID: 1,
			Items: []CreateProductComponentItemParams{
				{ComponentProductItemID: 2, Quantity: 1, UnitID: 1},
			},
		})
		if err != domain.ErrInvalidProductComponent {
			t.Errorf("expected ErrInvalidProductComponent, got %v", err)
		}
	})

	t.Run("missing product_item_id returns error", func(t *testing.T) {
		uc := NewProductComponentUsecase(&mockComponentRepo{})
		_, err := uc.Create(ctx, CreateProductComponentParams{
			MerchantID: 1,
			Items: []CreateProductComponentItemParams{
				{ComponentProductItemID: 2, Quantity: 1, UnitID: 1},
			},
		})
		if err != domain.ErrInvalidProductComponent {
			t.Errorf("expected ErrInvalidProductComponent, got %v", err)
		}
	})

	t.Run("empty items returns error", func(t *testing.T) {
		uc := NewProductComponentUsecase(&mockComponentRepo{})
		_, err := uc.Create(ctx, CreateProductComponentParams{
			MerchantID:    1,
			ProductItemID: 1,
			Items:         []CreateProductComponentItemParams{},
		})
		if err != domain.ErrNoComponentItems {
			t.Errorf("expected ErrNoComponentItems, got %v", err)
		}
	})

	t.Run("duplicate component for same product item rejected", func(t *testing.T) {
		repo := &mockComponentRepo{}
		uc := NewProductComponentUsecase(repo)

		uc.Create(ctx, CreateProductComponentParams{
			MerchantID:    1,
			ProductItemID: 1,
			Items:         []CreateProductComponentItemParams{{ComponentProductItemID: 2, Quantity: 1, UnitID: 1}},
		})

		_, err := uc.Create(ctx, CreateProductComponentParams{
			MerchantID:    1,
			ProductItemID: 1,
			Items:         []CreateProductComponentItemParams{{ComponentProductItemID: 3, Quantity: 2, UnitID: 1}},
		})
		if err != domain.ErrComponentAlreadyExists {
			t.Errorf("expected ErrComponentAlreadyExists, got %v", err)
		}
	})

	t.Run("different merchant can have component for same product item", func(t *testing.T) {
		repo := &mockComponentRepo{}
		uc := NewProductComponentUsecase(repo)

		uc.Create(ctx, CreateProductComponentParams{
			MerchantID:    1,
			ProductItemID: 1,
			Items:         []CreateProductComponentItemParams{{ComponentProductItemID: 2, Quantity: 1, UnitID: 1}},
		})

		_, err := uc.Create(ctx, CreateProductComponentParams{
			MerchantID:    2,
			ProductItemID: 1,
			Items:         []CreateProductComponentItemParams{{ComponentProductItemID: 3, Quantity: 2, UnitID: 1}},
		})
		if err != nil {
			t.Fatalf("expected success for different merchant, got %v", err)
		}
	})
}

func TestProductComponentUsecase_GetByProductItem(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("returns component", func(t *testing.T) {
		repo := &mockComponentRepo{}
		uc := NewProductComponentUsecase(repo)
		uc.Create(ctx, CreateProductComponentParams{
			MerchantID:    1,
			ProductItemID: 1,
			Items:         []CreateProductComponentItemParams{{ComponentProductItemID: 2, Quantity: 1, UnitID: 1}},
		})

		component, err := uc.GetByProductItem(ctx, 1, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if component.ProductItemID != 1 {
			t.Errorf("expected productItemID 1, got %d", component.ProductItemID)
		}
	})

	t.Run("not found returns error", func(t *testing.T) {
		uc := NewProductComponentUsecase(&mockComponentRepo{})
		_, err := uc.GetByProductItem(ctx, 999, 1)
		if err != domain.ErrProductComponentNotFound {
			t.Errorf("expected ErrProductComponentNotFound, got %v", err)
		}
	})
}

func TestProductComponentUsecase_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("replaces items", func(t *testing.T) {
		repo := &mockComponentRepo{}
		uc := NewProductComponentUsecase(repo)
		uc.Create(ctx, CreateProductComponentParams{
			MerchantID:    1,
			ProductItemID: 1,
			Items:         []CreateProductComponentItemParams{{ComponentProductItemID: 2, Quantity: 1, UnitID: 1}},
		})

		component, err := uc.Update(ctx, UpdateProductComponentParams{
			MerchantID:    1,
			ProductItemID: 1,
			Items: []CreateProductComponentItemParams{
				{ComponentProductItemID: 3, Quantity: 3, UnitID: 2},
				{ComponentProductItemID: 4, Quantity: 2, UnitID: 2},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(component.Items) != 2 {
			t.Errorf("expected 2 items, got %d", len(component.Items))
		}
	})
}

func TestProductComponentUsecase_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("deletes component", func(t *testing.T) {
		repo := &mockComponentRepo{}
		uc := NewProductComponentUsecase(repo)
		uc.Create(ctx, CreateProductComponentParams{
			MerchantID:    1,
			ProductItemID: 1,
			Items:         []CreateProductComponentItemParams{{ComponentProductItemID: 2, Quantity: 1, UnitID: 1}},
		})

		err := uc.Delete(ctx, 1, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		_, err = uc.GetByProductItem(ctx, 1, 1)
		if err != domain.ErrProductComponentNotFound {
			t.Errorf("expected ErrProductComponentNotFound after delete, got %v", err)
		}
	})

	t.Run("cross-merchant delete blocked", func(t *testing.T) {
		repo := &mockComponentRepo{}
		uc := NewProductComponentUsecase(repo)
		uc.Create(ctx, CreateProductComponentParams{
			MerchantID:    1,
			ProductItemID: 1,
			Items:         []CreateProductComponentItemParams{{ComponentProductItemID: 2, Quantity: 1, UnitID: 1}},
		})

		err := uc.Delete(ctx, 1, 2)
		if err != domain.ErrProductComponentNotFound {
			t.Errorf("expected ErrProductComponentNotFound for cross-merchant delete, got %v", err)
		}
	})
}

func TestProductComponentUsecase_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("lists components for merchant", func(t *testing.T) {
		repo := &mockComponentRepo{}
		uc := NewProductComponentUsecase(repo)
		uc.Create(ctx, CreateProductComponentParams{
			MerchantID:    1,
			ProductItemID: 1,
			Items:         []CreateProductComponentItemParams{{ComponentProductItemID: 2, Quantity: 1, UnitID: 1}},
		})
		uc.Create(ctx, CreateProductComponentParams{
			MerchantID:    1,
			ProductItemID: 2,
			Items:         []CreateProductComponentItemParams{{ComponentProductItemID: 3, Quantity: 2, UnitID: 2}},
		})
		uc.Create(ctx, CreateProductComponentParams{
			MerchantID:    2,
			ProductItemID: 3,
			Items:         []CreateProductComponentItemParams{{ComponentProductItemID: 4, Quantity: 1, UnitID: 1}},
		})

		components, err := uc.List(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(components) != 2 {
			t.Errorf("expected 2 components for merchant 1, got %d", len(components))
		}
	})
}
