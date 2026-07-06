package memory

import (
	"context"
	"testing"

	"github.com/saleforge/pos/services/internal/merchant/domain"
)

func TestStaffRepository_CreateAndGetByID(t *testing.T) {
	t.Parallel()
	repo := NewStaffRepository()
	staff := &domain.StaffMember{
		ID:         "s1",
		MerchantID: "m1",
		BranchID:   "b1",
		UserID:     "u1",
		Role:       domain.StaffRoleCashier,
		Status:     domain.StaffStatusActive,
	}

	err := repo.Create(context.Background(), staff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.GetByID(context.Background(), "s1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.UserID != "u1" {
		t.Errorf("expected u1, got %s", got.UserID)
	}
	if got.Role != domain.StaffRoleCashier {
		t.Errorf("expected cashier, got %s", got.Role)
	}
}

func TestStaffRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	repo := NewStaffRepository()
	_, err := repo.GetByID(context.Background(), "nonexistent")
	if err != domain.ErrStaffNotFound {
		t.Errorf("expected ErrStaffNotFound, got %v", err)
	}
}

func TestStaffRepository_ListByBranch(t *testing.T) {
	t.Parallel()
	repo := NewStaffRepository()
	repo.Create(context.Background(), &domain.StaffMember{ID: "s1", BranchID: "b1"})
	repo.Create(context.Background(), &domain.StaffMember{ID: "s2", BranchID: "b1"})
	repo.Create(context.Background(), &domain.StaffMember{ID: "s3", BranchID: "b2"})

	staff, err := repo.ListByBranch(context.Background(), "b1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(staff) != 2 {
		t.Errorf("expected 2 staff, got %d", len(staff))
	}

	empty, err := repo.ListByBranch(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("expected 0 staff, got %d", len(empty))
	}
}

func TestStaffRepository_ListByMerchant(t *testing.T) {
	t.Parallel()
	repo := NewStaffRepository()
	repo.Create(context.Background(), &domain.StaffMember{ID: "s1", MerchantID: "m1"})
	repo.Create(context.Background(), &domain.StaffMember{ID: "s2", MerchantID: "m1"})
	repo.Create(context.Background(), &domain.StaffMember{ID: "s3", MerchantID: "m2"})

	staff, err := repo.ListByMerchant(context.Background(), "m1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(staff) != 2 {
		t.Errorf("expected 2 staff, got %d", len(staff))
	}
}

func TestStaffRepository_GetByUserAndBranch(t *testing.T) {
	t.Parallel()
	repo := NewStaffRepository()
	repo.Create(context.Background(), &domain.StaffMember{ID: "s1", UserID: "u1", BranchID: "b1"})
	repo.Create(context.Background(), &domain.StaffMember{ID: "s2", UserID: "u2", BranchID: "b1"})

	got, err := repo.GetByUserAndBranch(context.Background(), "u1", "b1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != "s1" {
		t.Errorf("expected s1, got %s", got.ID)
	}

	_, err = repo.GetByUserAndBranch(context.Background(), "u1", "b2")
	if err != domain.ErrStaffNotFound {
		t.Errorf("expected ErrStaffNotFound, got %v", err)
	}
}

func TestStaffRepository_Update(t *testing.T) {
	t.Parallel()
	repo := NewStaffRepository()
	repo.Create(context.Background(), &domain.StaffMember{ID: "s1", Role: domain.StaffRoleCashier})

	err := repo.Update(context.Background(), &domain.StaffMember{ID: "s1", Role: domain.StaffRoleManager})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := repo.GetByID(context.Background(), "s1")
	if got.Role != domain.StaffRoleManager {
		t.Errorf("expected manager, got %s", got.Role)
	}
}

func TestStaffRepository_Delete(t *testing.T) {
	t.Parallel()
	repo := NewStaffRepository()
	repo.Create(context.Background(), &domain.StaffMember{ID: "s1"})

	err := repo.Delete(context.Background(), "s1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = repo.GetByID(context.Background(), "s1")
	if err != domain.ErrStaffNotFound {
		t.Errorf("expected ErrStaffNotFound, got %v", err)
	}
}
