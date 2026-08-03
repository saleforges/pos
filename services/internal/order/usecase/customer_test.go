package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/saleforge/pos/services/internal/order/domain"
)

func TestCustomerUsecase_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uc := NewCustomerUsecase(&mockCustomerRepo{})
		c, err := uc.Create(ctx, CreateCustomerParams{MerchantID: 1, Name: "Pak Budi", Phone: "0812345"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.ID == 0 {
			t.Error("expected non-zero id")
		}
		if c.Name != "Pak Budi" {
			t.Errorf("expected name 'Pak Budi', got '%s'", c.Name)
		}
	})

	t.Run("empty name rejected", func(t *testing.T) {
		uc := NewCustomerUsecase(&mockCustomerRepo{})
		_, err := uc.Create(ctx, CreateCustomerParams{MerchantID: 1})
		if err != domain.ErrInvalidCustomer {
			t.Errorf("expected ErrInvalidCustomer, got %v", err)
		}
	})
}

func TestCustomerUsecase_GetByID(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("cross-merchant returns not found", func(t *testing.T) {
		repo := &mockCustomerRepo{}
		uc := NewCustomerUsecase(repo)
		uc.Create(ctx, CreateCustomerParams{MerchantID: 1, Name: "Pak Budi"})

		_, err := uc.GetByID(ctx, 1, 2)
		if err != domain.ErrCustomerNotFound {
			t.Errorf("expected ErrCustomerNotFound, got %v", err)
		}
	})
}

func TestCustomerUsecase_Update(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("partial update", func(t *testing.T) {
		repo := &mockCustomerRepo{}
		uc := NewCustomerUsecase(repo)
		uc.Create(ctx, CreateCustomerParams{MerchantID: 1, Name: "Pak Budi"})

		phone := "081999"
		c, err := uc.Update(ctx, UpdateCustomerParams{ID: 1, MerchantID: 1, Phone: &phone})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if c.Phone != "081999" {
			t.Errorf("expected phone '081999', got '%s'", c.Phone)
		}
		if c.Name != "Pak Budi" {
			t.Errorf("expected name preserved, got '%s'", c.Name)
		}
	})

	t.Run("empty name on update rejected", func(t *testing.T) {
		repo := &mockCustomerRepo{}
		uc := NewCustomerUsecase(repo)
		uc.Create(ctx, CreateCustomerParams{MerchantID: 1, Name: "Pak Budi"})

		name := ""
		_, err := uc.Update(ctx, UpdateCustomerParams{ID: 1, MerchantID: 1, Name: &name})
		if err != domain.ErrInvalidCustomer {
			t.Errorf("expected ErrInvalidCustomer, got %v", err)
		}
	})
}

func TestCustomerUsecase_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("cross-merchant delete blocked", func(t *testing.T) {
		repo := &mockCustomerRepo{}
		uc := NewCustomerUsecase(repo)
		uc.Create(ctx, CreateCustomerParams{MerchantID: 1, Name: "Pak Budi"})

		err := uc.Delete(ctx, 1, 2)
		if err != domain.ErrCustomerNotFound {
			t.Errorf("expected ErrCustomerNotFound, got %v", err)
		}
	})

	t.Run("own customer delete succeeds", func(t *testing.T) {
		repo := &mockCustomerRepo{}
		uc := NewCustomerUsecase(repo)
		uc.Create(ctx, CreateCustomerParams{MerchantID: 1, Name: "Pak Budi"})

		if err := uc.Delete(ctx, 1, 1); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err := uc.GetByID(ctx, 1, 1)
		if err != domain.ErrCustomerNotFound {
			t.Errorf("expected ErrCustomerNotFound after delete, got %v", err)
		}
	})
}

func TestCustomerUsecase_Sync(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("returns changes since lastSync with token", func(t *testing.T) {
		repo := &mockCustomerRepo{}
		uc := NewCustomerUsecase(repo)
		now := time.Now().UTC()
		repo.customers = map[int64]*domain.Customer{
			1: {ID: 1, MerchantID: 1, Name: "Pak Budi", Phone: "0811", CreatedAt: now, UpdatedAt: now},
			2: {ID: 2, MerchantID: 2, Name: "Orang Lain", Phone: "0822", CreatedAt: now, UpdatedAt: now},
		}

		since := now.Add(-time.Hour)
		result, err := uc.Sync(ctx, 1, &since)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Customers) != 1 {
			t.Errorf("expected 1 customer for merchant 1, got %d", len(result.Customers))
		}
		if result.SyncToken == "" {
			t.Error("expected non-empty sync token")
		}
	})

	t.Run("no lastSync returns all", func(t *testing.T) {
		repo := &mockCustomerRepo{}
		uc := NewCustomerUsecase(repo)
		now := time.Now().UTC()
		repo.customers = map[int64]*domain.Customer{
			1: {ID: 1, MerchantID: 1, Name: "A", CreatedAt: now, UpdatedAt: now},
			2: {ID: 2, MerchantID: 1, Name: "B", CreatedAt: now, UpdatedAt: now},
		}

		result, err := uc.Sync(ctx, 1, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Customers) != 2 {
			t.Errorf("expected 2 customers, got %d", len(result.Customers))
		}
	})
}
