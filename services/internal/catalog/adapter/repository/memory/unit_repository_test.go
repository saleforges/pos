package memory

import (
	"context"
	"testing"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

func TestUnitRepository(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("get all returns seeded units", func(t *testing.T) {
		repo := NewUnitRepository()
		units, err := repo.GetAll(ctx)
		if err != nil {
			t.Fatalf("GetAll: %v", err)
		}
		if len(units) != 8 {
			t.Errorf("expected 8 units, got %d", len(units))
		}
	})

	t.Run("get by id", func(t *testing.T) {
		repo := NewUnitRepository()
		u, err := repo.GetByID(ctx, 1)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if u.Code != "PCS" {
			t.Errorf("expected 'PCS', got '%s'", u.Code)
		}
	})

	t.Run("get by id not found", func(t *testing.T) {
		repo := NewUnitRepository()
		_, err := repo.GetByID(ctx, 999)
		if err != domain.ErrUnitNotFound {
			t.Errorf("expected ErrUnitNotFound, got %v", err)
		}
	})

	t.Run("get by code", func(t *testing.T) {
		repo := NewUnitRepository()
		u, err := repo.GetByCode(ctx, "KG")
		if err != nil {
			t.Fatalf("GetByCode: %v", err)
		}
		if u.Name != "Kilogram" {
			t.Errorf("expected 'Kilogram', got '%s'", u.Name)
		}
	})
}
