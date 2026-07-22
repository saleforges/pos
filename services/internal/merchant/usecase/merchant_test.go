package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/pkg/pagination"
)

func TestMerchantUsecase_CreateMerchant(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        CreateMerchantParams
		merchantRepo *mockMerchantRepo
		wantErr      error
	}{
		{
			name: "successful creation with defaults",
			input: CreateMerchantParams{
				Name:  "Test Merchant",
				Email: "test@example.com",
			},
			merchantRepo: &mockMerchantRepo{},
		},
		{
			name: "missing name",
			input: CreateMerchantParams{
				Email: "test@example.com",
			},
			merchantRepo: &mockMerchantRepo{},
			wantErr:      domain.ErrInvalidMerchant,
		},
		{
			name: "missing email",
			input: CreateMerchantParams{
				Name: "No Email",
			},
			merchantRepo: &mockMerchantRepo{},
			wantErr:      domain.ErrInvalidMerchant,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := NewMerchantUsecase(tt.merchantRepo, &mockBranchRepo{}, &mockStaffRepo{})
			result, err := uc.CreateMerchant(context.Background(), tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected result, got nil")
			}
			if result.Name != tt.input.Name {
				t.Errorf("expected name %s, got %s", tt.input.Name, result.Name)
			}
			if result.Status != domain.MerchantStatusActive {
				t.Errorf("expected status active, got %s", result.Status)
			}
			if result.Settings.Currency != "IDR" {
				t.Errorf("expected default currency IDR, got %s", result.Settings.Currency)
			}
			if result.Settings.Timezone != "Asia/Jakarta" {
				t.Errorf("expected default timezone Asia/Jakarta, got %s", result.Settings.Timezone)
			}
		})
	}
}

func TestMerchantUsecase_GetMerchant(t *testing.T) {
	t.Parallel()

	uc := NewMerchantUsecase(&mockMerchantRepo{}, &mockBranchRepo{}, &mockStaffRepo{})
	_, err := uc.GetMerchant(context.Background(), 999)
	if err != domain.ErrMerchantNotFound {
		t.Errorf("expected ErrMerchantNotFound, got %v", err)
	}
}

func TestMerchantUsecase_CreateBranch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        CreateBranchParams
		merchantRepo *mockMerchantRepo
		branchRepo   *mockBranchRepo
		wantErr      error
	}{
		{
			name: "successful branch creation",
			input: CreateBranchParams{
				MerchantID: 1,
				Name:       "Branch A",
				Code:       "BR-A",
			},
			merchantRepo: &mockMerchantRepo{merchants: map[int64]*domain.Merchant{1: {ID: 1, Name: "M1"}}},
			branchRepo:   &mockBranchRepo{},
		},
		{
			name: "merchant not found",
			input: CreateBranchParams{
				MerchantID: 999,
				Name:       "Branch B",
				Code:       "BR-B",
			},
			merchantRepo: &mockMerchantRepo{},
			branchRepo:   &mockBranchRepo{},
			wantErr:      domain.ErrMerchantNotFound,
		},
		{
			name: "missing required fields",
			input: CreateBranchParams{
				MerchantID: 1,
			},
			merchantRepo: &mockMerchantRepo{merchants: map[int64]*domain.Merchant{1: {ID: 1}}},
			branchRepo:   &mockBranchRepo{},
			wantErr:      domain.ErrInvalidBranch,
		},
		{
			name: "branch code already exists",
			input: CreateBranchParams{
				MerchantID: 1,
				Name:       "Branch C",
				Code:       "BR-C",
			},
			merchantRepo: &mockMerchantRepo{merchants: map[int64]*domain.Merchant{1: {ID: 1}}},
			branchRepo:   &mockBranchRepo{err: domain.ErrBranchExists},
			wantErr:      domain.ErrBranchExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := NewMerchantUsecase(tt.merchantRepo, tt.branchRepo, &mockStaffRepo{})
			result, err := uc.CreateBranch(context.Background(), tt.input)

			if tt.wantErr != nil {
				if err == nil { t.Fatal("expected error, got nil") }
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Name != tt.input.Name {
				t.Errorf("expected name %s, got %s", tt.input.Name, result.Name)
			}
		})
	}
}

func TestMerchantUsecase_AssignStaff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        AssignStaffParams
		merchantRepo *mockMerchantRepo
		branchRepo   *mockBranchRepo
		staffRepo    *mockStaffRepo
		wantErr      error
	}{
		{
			name: "successful assignment",
			input: AssignStaffParams{
				MerchantID: 1,
				BranchID:   1,
				UserID:     1,
				Role:       domain.StaffRoleCashier,
			},
			merchantRepo: &mockMerchantRepo{merchants: map[int64]*domain.Merchant{1: {ID: 1}}},
			branchRepo:   &mockBranchRepo{branches: map[int64]*domain.Branch{1: {ID: 1, MerchantID: 1}}},
			staffRepo:    &mockStaffRepo{},
		},
		{
			name: "merchant not found",
			input: AssignStaffParams{
				MerchantID: 999,
				BranchID:   1,
				UserID:     1,
				Role:       domain.StaffRoleCashier,
			},
			merchantRepo: &mockMerchantRepo{},
			branchRepo:   &mockBranchRepo{branches: map[int64]*domain.Branch{1: {ID: 1}}},
			staffRepo:    &mockStaffRepo{},
			wantErr:      domain.ErrMerchantNotFound,
		},
		{
			name: "branch not found",
			input: AssignStaffParams{
				MerchantID: 1,
				BranchID:   999,
				UserID:     1,
				Role:       domain.StaffRoleCashier,
			},
			merchantRepo: &mockMerchantRepo{merchants: map[int64]*domain.Merchant{1: {ID: 1}}},
			branchRepo:   &mockBranchRepo{},
			staffRepo:    &mockStaffRepo{},
			wantErr:      domain.ErrBranchNotFound,
		},
		{
			name: "duplicate assignment",
			input: AssignStaffParams{
				MerchantID: 1,
				BranchID:   1,
				UserID:     1,
				Role:       domain.StaffRoleCashier,
			},
			merchantRepo: &mockMerchantRepo{merchants: map[int64]*domain.Merchant{1: {ID: 1}}},
			branchRepo:   &mockBranchRepo{branches: map[int64]*domain.Branch{1: {ID: 1, MerchantID: 1}}},
			staffRepo:    &mockStaffRepo{staff: map[int64]*domain.StaffMember{1: {UserID: 1, BranchID: 1}}},
			wantErr:      domain.ErrStaffExists,
		},
		{
			name: "missing input fields",
			input: AssignStaffParams{
				MerchantID: 1,
				BranchID:   1,
			},
			merchantRepo: &mockMerchantRepo{merchants: map[int64]*domain.Merchant{1: {ID: 1}}},
			branchRepo:   &mockBranchRepo{branches: map[int64]*domain.Branch{1: {ID: 1}}},
			staffRepo:    &mockStaffRepo{},
			wantErr:      domain.ErrInvalidStaff,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			uc := NewMerchantUsecase(tt.merchantRepo, tt.branchRepo, tt.staffRepo)
			result, err := uc.AssignStaff(context.Background(), tt.input)

			if tt.wantErr != nil {
				if err == nil { t.Fatal("expected error, got nil") }
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected result, got nil")
			}
		})
	}
}

func TestMerchantUsecase_ListMerchants(t *testing.T) {
	t.Parallel()

	repo := &mockMerchantRepo{
		merchants: map[int64]*domain.Merchant{
			1: {ID: 1, Name: "M1", Email: "m1@t.com"},
			2: {ID: 2, Name: "M2", Email: "m2@t.com"},
		},
	}
	uc := NewMerchantUsecase(repo, &mockBranchRepo{}, &mockStaffRepo{})

	data, meta, err := uc.ListMerchants(context.Background(), pagination.Params{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.Total != 2 {
		t.Errorf("expected total 2, got %d", meta.Total)
	}
	if len(data) != 2 {
		t.Errorf("expected 2 merchants, got %d", len(data))
	}
}

func TestMerchantUsecase_ListBranches(t *testing.T) {
	t.Parallel()

	branchRepo := &mockBranchRepo{
		branches: map[int64]*domain.Branch{
			1: {ID: 1, MerchantID: 1, Name: "B1"},
			2: {ID: 2, MerchantID: 1, Name: "B2"},
		},
	}
	uc := NewMerchantUsecase(&mockMerchantRepo{}, branchRepo, &mockStaffRepo{})

	data, meta, err := uc.ListBranches(context.Background(), 1, pagination.Params{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.Total != 2 {
		t.Errorf("expected total 2, got %d", meta.Total)
	}
	if len(data) != 2 {
		t.Errorf("expected 2 branches, got %d", len(data))
	}
}

func TestMerchantUsecase_UpdateMerchant(t *testing.T) {
	t.Parallel()

	username := "Updated Name"
	uc := NewMerchantUsecase(&mockMerchantRepo{
		merchants: map[int64]*domain.Merchant{1: {ID: 1, Name: "Old", Email: "old@t.com"}},
	}, &mockBranchRepo{}, &mockStaffRepo{})

	merchant, err := uc.UpdateMerchant(context.Background(), UpdateMerchantParams{
		ID: 1, Name: &username,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if merchant.Name != "Updated Name" {
		t.Errorf("expected 'Updated Name', got %s", merchant.Name)
	}
}

func TestMerchantUsecase_DeleteMerchant(t *testing.T) {
	t.Parallel()

	repo := &mockMerchantRepo{
		merchants: map[int64]*domain.Merchant{1: {ID: 1, Name: "M1", Email: "m1@t.com"}},
	}
	uc := NewMerchantUsecase(repo, &mockBranchRepo{}, &mockStaffRepo{})

	if err := uc.DeleteMerchant(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := uc.GetMerchant(context.Background(), 1)
	if err != domain.ErrMerchantNotFound {
		t.Error("expected merchant to be deleted")
	}
}
