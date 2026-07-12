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
		MerchantID: 1,
		BranchID:   1,
		UserID:     1,
		Role:       domain.StaffRoleCashier,
		Status:     domain.StaffStatusActive,
	}

	err := repo.Create(context.Background(), staff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if staff.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}

	got, err := repo.GetByID(context.Background(), staff.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.UserID != 1 {
		t.Errorf("expected user ID 1, got %d", got.UserID)
	}
	if got.Role != domain.StaffRoleCashier {
		t.Errorf("expected cashier, got %s", got.Role)
	}
}

func TestStaffRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	repo := NewStaffRepository()
	_, err := repo.GetByID(context.Background(), 999)
	if err != domain.ErrStaffNotFound {
		t.Errorf("expected ErrStaffNotFound, got %v", err)
	}
}

func TestStaffRepository_ListByBranch(t *testing.T) {
	t.Parallel()
	repo := NewStaffRepository()
	repo.Create(context.Background(), &domain.StaffMember{BranchID: 1})
	repo.Create(context.Background(), &domain.StaffMember{BranchID: 1})
	repo.Create(context.Background(), &domain.StaffMember{BranchID: 2})

	staff, err := repo.ListByBranch(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(staff) != 2 {
		t.Errorf("expected 2 staff, got %d", len(staff))
	}

	empty, err := repo.ListByBranch(context.Background(), 999)
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
	repo.Create(context.Background(), &domain.StaffMember{MerchantID: 1})
	repo.Create(context.Background(), &domain.StaffMember{MerchantID: 1})
	repo.Create(context.Background(), &domain.StaffMember{MerchantID: 2})

	staff, err := repo.ListByMerchant(context.Background(), 1)
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
	repo.Create(context.Background(), &domain.StaffMember{UserID: 1, BranchID: 1})
	repo.Create(context.Background(), &domain.StaffMember{UserID: 2, BranchID: 1})

	got, err := repo.GetByUserAndBranch(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.UserID != 1 {
		t.Errorf("expected user ID 1, got %d", got.UserID)
	}

	_, err = repo.GetByUserAndBranch(context.Background(), 1, 2)
	if err != domain.ErrStaffNotFound {
		t.Errorf("expected ErrStaffNotFound, got %v", err)
	}
}

func TestStaffRepository_Update(t *testing.T) {
	t.Parallel()
	repo := NewStaffRepository()
	s := &domain.StaffMember{Role: domain.StaffRoleCashier}
	repo.Create(context.Background(), s)

	err := repo.Update(context.Background(), &domain.StaffMember{ID: s.ID, Role: domain.StaffRoleManager})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := repo.GetByID(context.Background(), s.ID)
	if got.Role != domain.StaffRoleManager {
		t.Errorf("expected manager, got %s", got.Role)
	}
}

func TestStaffRepository_Delete(t *testing.T) {
	t.Parallel()
	repo := NewStaffRepository()
	s := &domain.StaffMember{}
	repo.Create(context.Background(), s)

	err := repo.Delete(context.Background(), s.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = repo.GetByID(context.Background(), s.ID)
	if err != domain.ErrStaffNotFound {
		t.Errorf("expected ErrStaffNotFound, got %v", err)
	}
}
