package memory

import (
	"context"
	"testing"

	"github.com/saleforge/pos/services/internal/order/domain"
)

func TestCustomerRepository(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("create and get by id", func(t *testing.T) {
		repo := NewCustomerRepository()
		c := &domain.Customer{MerchantID: 1, Name: "Pak Budi", Phone: "0812345"}
		if err := repo.Create(ctx, c); err != nil {
			t.Fatalf("Create: %v", err)
		}
		if c.ID == 0 {
			t.Error("expected non-zero id")
		}
		got, err := repo.GetByID(ctx, c.ID, 1)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if got.Name != "Pak Budi" {
			t.Errorf("expected 'Pak Budi', got '%s'", got.Name)
		}
	})

	t.Run("cross-merchant get returns not found", func(t *testing.T) {
		repo := NewCustomerRepository()
		repo.Create(ctx, &domain.Customer{MerchantID: 1, Name: "Pak Budi"})
		_, err := repo.GetByID(ctx, 1, 2)
		if err != domain.ErrCustomerNotFound {
			t.Errorf("expected ErrCustomerNotFound, got %v", err)
		}
	})

	t.Run("list with search", func(t *testing.T) {
		repo := NewCustomerRepository()
		repo.Create(ctx, &domain.Customer{MerchantID: 1, Name: "Pak Budi", Phone: "0812345"})
		repo.Create(ctx, &domain.Customer{MerchantID: 1, Name: "Bu Sari", Phone: "082111"})
		repo.Create(ctx, &domain.Customer{MerchantID: 2, Name: "Pak Lain", Phone: "083333"})

		all, _ := repo.List(ctx, 1, "")
		if len(all) != 2 {
			t.Errorf("expected 2 customers for merchant 1, got %d", len(all))
		}
		searched, _ := repo.List(ctx, 1, "budi")
		if len(searched) != 1 {
			t.Errorf("expected 1 customer for search 'budi', got %d", len(searched))
		}
		byPhone, _ := repo.List(ctx, 1, "082111")
		if len(byPhone) != 1 {
			t.Errorf("expected 1 customer for phone search, got %d", len(byPhone))
		}
	})

	t.Run("update", func(t *testing.T) {
		repo := NewCustomerRepository()
		repo.Create(ctx, &domain.Customer{MerchantID: 1, Name: "Pak Budi"})
		c := &domain.Customer{ID: 1, MerchantID: 1, Name: "Pak Budi Baru"}
		if err := repo.Update(ctx, c); err != nil {
			t.Fatalf("Update: %v", err)
		}
		got, _ := repo.GetByID(ctx, 1, 1)
		if got.Name != "Pak Budi Baru" {
			t.Errorf("expected 'Pak Budi Baru', got '%s'", got.Name)
		}
	})

	t.Run("cross-merchant update blocked", func(t *testing.T) {
		repo := NewCustomerRepository()
		repo.Create(ctx, &domain.Customer{MerchantID: 1, Name: "Pak Budi"})
		err := repo.Update(ctx, &domain.Customer{ID: 1, MerchantID: 2, Name: "Hacked"})
		if err != domain.ErrCustomerNotFound {
			t.Errorf("expected ErrCustomerNotFound, got %v", err)
		}
	})

	t.Run("delete", func(t *testing.T) {
		repo := NewCustomerRepository()
		repo.Create(ctx, &domain.Customer{MerchantID: 1, Name: "Pak Budi"})
		if err := repo.Delete(ctx, 1, 1); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		_, err := repo.GetByID(ctx, 1, 1)
		if err != domain.ErrCustomerNotFound {
			t.Errorf("expected ErrCustomerNotFound after delete, got %v", err)
		}
	})
}
