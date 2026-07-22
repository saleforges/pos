package memory

import (
	"context"
	"testing"

	"github.com/saleforge/pos/services/internal/merchant/domain"
)

func TestMerchantRepository_CreateAndGetByID(t *testing.T) {
	t.Parallel()
	repo := NewMerchantRepository()
	m := &domain.Merchant{Name: "Test Merchant", Email: "test@test.com"}

	err := repo.Create(context.Background(), m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}

	got, err := repo.GetByID(context.Background(), m.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Test Merchant" {
		t.Errorf("expected Test Merchant, got %s", got.Name)
	}
}

func TestMerchantRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	repo := NewMerchantRepository()
	_, err := repo.GetByID(context.Background(), 999)
	if err != domain.ErrMerchantNotFound {
		t.Errorf("expected ErrMerchantNotFound, got %v", err)
	}
}

func TestMerchantRepository_List(t *testing.T) {
	t.Parallel()
	repo := NewMerchantRepository()
	repo.Create(context.Background(), &domain.Merchant{Name: "A"})
	repo.Create(context.Background(), &domain.Merchant{Name: "B"})
	repo.Create(context.Background(), &domain.Merchant{Name: "C"})

	all, _, err := repo.List(context.Background(), 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 merchants, got %d", len(all))
	}
}

func TestMerchantRepository_Update(t *testing.T) {
	t.Parallel()
	repo := NewMerchantRepository()
	m := &domain.Merchant{Name: "Old"}
	repo.Create(context.Background(), m)

	err := repo.Update(context.Background(), &domain.Merchant{ID: m.ID, Name: "New"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := repo.GetByID(context.Background(), m.ID)
	if got.Name != "New" {
		t.Errorf("expected New, got %s", got.Name)
	}
}

func TestMerchantRepository_Delete(t *testing.T) {
	t.Parallel()
	repo := NewMerchantRepository()
	m := &domain.Merchant{Name: "Test"}
	repo.Create(context.Background(), m)

	err := repo.Delete(context.Background(), m.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = repo.GetByID(context.Background(), m.ID)
	if err != domain.ErrMerchantNotFound {
		t.Errorf("expected ErrMerchantNotFound, got %v", err)
	}
}

func TestBranchRepository_CreateAndGetByID(t *testing.T) {
	t.Parallel()
	repo := NewBranchRepository()
	b := &domain.Branch{Name: "Branch A", Code: "BR-A"}

	err := repo.Create(context.Background(), b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}

	got, err := repo.GetByID(context.Background(), b.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Code != "BR-A" {
		t.Errorf("expected BR-A, got %s", got.Code)
	}
}

func TestBranchRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	repo := NewBranchRepository()
	_, err := repo.GetByID(context.Background(), 999)
	if err != domain.ErrBranchNotFound {
		t.Errorf("expected ErrBranchNotFound, got %v", err)
	}
}

func TestBranchRepository_ListByMerchant(t *testing.T) {
	t.Parallel()
	repo := NewBranchRepository()
	repo.Create(context.Background(), &domain.Branch{MerchantID: 1})
	repo.Create(context.Background(), &domain.Branch{MerchantID: 1})
	repo.Create(context.Background(), &domain.Branch{MerchantID: 2})

	branches, _, err := repo.ListByMerchant(context.Background(), 1, 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(branches) != 2 {
		t.Errorf("expected 2 branches, got %d", len(branches))
	}

	empty, _, err := repo.ListByMerchant(context.Background(), 999, 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("expected 0 branches, got %d", len(empty))
	}
}

func TestBranchRepository_Update(t *testing.T) {
	t.Parallel()
	repo := NewBranchRepository()
	b := &domain.Branch{Name: "Old"}
	repo.Create(context.Background(), b)

	err := repo.Update(context.Background(), &domain.Branch{ID: b.ID, Name: "New"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := repo.GetByID(context.Background(), b.ID)
	if got.Name != "New" {
		t.Errorf("expected New, got %s", got.Name)
	}
}

func TestBranchRepository_Delete(t *testing.T) {
	t.Parallel()
	repo := NewBranchRepository()
	b := &domain.Branch{Name: "Test"}
	repo.Create(context.Background(), b)

	err := repo.Delete(context.Background(), b.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = repo.GetByID(context.Background(), b.ID)
	if err != domain.ErrBranchNotFound {
		t.Errorf("expected ErrBranchNotFound, got %v", err)
	}
}
