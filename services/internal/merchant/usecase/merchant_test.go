package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/internal/merchant/port/repository"
)

type mockMerchantRepo struct {
	merchants map[int64]*domain.Merchant
	err       error
}

func (m *mockMerchantRepo) Create(_ context.Context, merchant *domain.Merchant) error {
	if m.err != nil {
		return m.err
	}
	if m.merchants == nil {
		m.merchants = make(map[int64]*domain.Merchant)
	}
	merchant.ID = int64(len(m.merchants) + 1)
	m.merchants[merchant.ID] = merchant
	return nil
}

func (m *mockMerchantRepo) GetByID(_ context.Context, id int64) (*domain.Merchant, error) {
	if m.err != nil {
		return nil, m.err
	}
	merchant, ok := m.merchants[id]
	if !ok {
		return nil, domain.ErrMerchantNotFound
	}
	return merchant, nil
}

func (m *mockMerchantRepo) List(_ context.Context, offset, limit int) ([]domain.Merchant, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.Merchant
	for _, merchant := range m.merchants {
		result = append(result, *merchant)
	}
	return result, nil
}

func (m *mockMerchantRepo) Update(_ context.Context, merchant *domain.Merchant) error {
	if m.err != nil {
		return m.err
	}
	m.merchants[merchant.ID] = merchant
	return nil
}

func (m *mockMerchantRepo) Delete(_ context.Context, id int64) error {
	if m.err != nil {
		return m.err
	}
	delete(m.merchants, id)
	return nil
}

type mockBranchRepo struct {
	branches map[int64]*domain.Branch
	err      error
}

func (m *mockBranchRepo) Create(_ context.Context, branch *domain.Branch) error {
	if m.err != nil {
		return m.err
	}
	if m.branches == nil {
		m.branches = make(map[int64]*domain.Branch)
	}
	branch.ID = int64(len(m.branches) + 1)
	m.branches[branch.ID] = branch
	return nil
}

func (m *mockBranchRepo) GetByID(_ context.Context, id int64) (*domain.Branch, error) {
	if m.err != nil {
		return nil, m.err
	}
	branch, ok := m.branches[id]
	if !ok {
		return nil, domain.ErrBranchNotFound
	}
	return branch, nil
}

func (m *mockBranchRepo) ListByMerchant(_ context.Context, merchantID int64) ([]domain.Branch, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.Branch
	for _, branch := range m.branches {
		if branch.MerchantID == merchantID {
			result = append(result, *branch)
		}
	}
	return result, nil
}

func (m *mockBranchRepo) Update(_ context.Context, branch *domain.Branch) error {
	if m.err != nil {
		return m.err
	}
	m.branches[branch.ID] = branch
	return nil
}

func (m *mockBranchRepo) Delete(_ context.Context, id int64) error {
	if m.err != nil {
		return m.err
	}
	delete(m.branches, id)
	return nil
}

type mockStaffRepo struct {
	staff map[int64]*domain.StaffMember
	err   error
}

func (m *mockStaffRepo) Create(_ context.Context, s *domain.StaffMember) error {
	if m.err != nil {
		return m.err
	}
	if m.staff == nil {
		m.staff = make(map[int64]*domain.StaffMember)
	}
	s.ID = int64(len(m.staff) + 1)
	m.staff[s.ID] = s
	return nil
}

func (m *mockStaffRepo) GetByID(_ context.Context, id int64) (*domain.StaffMember, error) {
	if m.err != nil {
		return nil, m.err
	}
	s, ok := m.staff[id]
	if !ok {
		return nil, domain.ErrStaffNotFound
	}
	return s, nil
}

func (m *mockStaffRepo) ListByBranch(_ context.Context, branchID int64) ([]domain.StaffMember, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.StaffMember
	for _, s := range m.staff {
		if s.BranchID == branchID {
			result = append(result, *s)
		}
	}
	return result, nil
}

func (m *mockStaffRepo) ListByMerchant(_ context.Context, merchantID int64) ([]domain.StaffMember, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.StaffMember
	for _, s := range m.staff {
		if s.MerchantID == merchantID {
			result = append(result, *s)
		}
	}
	return result, nil
}

func (m *mockStaffRepo) GetByUserAndBranch(_ context.Context, userID, branchID int64) (*domain.StaffMember, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, s := range m.staff {
		if s.UserID == userID && s.BranchID == branchID {
			return s, nil
		}
	}
	return nil, domain.ErrStaffNotFound
}

func (m *mockStaffRepo) ListByUserAndMerchant(_ context.Context, userID, merchantID int64) ([]domain.StaffMember, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []domain.StaffMember
	for _, s := range m.staff {
		if s.UserID == userID && s.MerchantID == merchantID {
			result = append(result, *s)
		}
	}
	if result == nil {
		return []domain.StaffMember{}, nil
	}
	return result, nil
}

func (m *mockStaffRepo) SetDefaultBranch(_ context.Context, userID, branchID int64) error {
	if m.err != nil {
		return m.err
	}
	for _, s := range m.staff {
		if s.UserID == userID {
			s.IsDefault = s.BranchID == branchID
		}
	}
	return nil
}

func (m *mockStaffRepo) Update(_ context.Context, s *domain.StaffMember) error {
	if m.err != nil {
		return m.err
	}
	m.staff[s.ID] = s
	return nil
}

func (m *mockStaffRepo) Delete(_ context.Context, id int64) error {
	if m.err != nil {
		return m.err
	}
	delete(m.staff, id)
	return nil
}

func TestMerchantUsecase_CreateMerchant(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        CreateMerchantInput
		merchantRepo repository.MerchantRepository
		branchRepo   repository.BranchRepository
		staffRepo    repository.StaffRepository
		wantErr      error
	}{
		{
			name: "successful creation with defaults",
			input: CreateMerchantInput{
				Name:  "Test Merchant",
				Email: "test@example.com",
			},
			merchantRepo: &mockMerchantRepo{},
			branchRepo:   &mockBranchRepo{},
			staffRepo:    &mockStaffRepo{},
		},
		{
			name: "missing name",
			input: CreateMerchantInput{
				Email: "test@example.com",
			},
			merchantRepo: &mockMerchantRepo{},
			branchRepo:   &mockBranchRepo{},
			staffRepo:    &mockStaffRepo{},
			wantErr:      domain.ErrInvalidMerchant,
		},
		{
			name: "missing email",
			input: CreateMerchantInput{
				Name: "No Email",
			},
			merchantRepo: &mockMerchantRepo{},
			branchRepo:   &mockBranchRepo{},
			staffRepo:    &mockStaffRepo{},
			wantErr:      domain.ErrInvalidMerchant,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := NewMerchantUsecase(tt.merchantRepo, tt.branchRepo, tt.staffRepo)
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
		setup        func(mRepo *mockMerchantRepo)
		input        CreateBranchInput
		merchantRepo repository.MerchantRepository
		branchRepo   repository.BranchRepository
		staffRepo    repository.StaffRepository
		wantErr      error
	}{
		{
			name: "successful branch creation",
			input: CreateBranchInput{
				MerchantID: 1,
				Name:       "Branch A",
				Code:       "BR-A",
			},
			merchantRepo: &mockMerchantRepo{merchants: map[int64]*domain.Merchant{1: {ID: 1, Name: "M1"}}},
			branchRepo:   &mockBranchRepo{},
			staffRepo:    &mockStaffRepo{},
		},
		{
			name: "merchant not found",
			input: CreateBranchInput{
				MerchantID: 999,
				Name:       "Branch B",
				Code:       "BR-B",
			},
			merchantRepo: &mockMerchantRepo{},
			branchRepo:   &mockBranchRepo{},
			staffRepo:    &mockStaffRepo{},
			wantErr:      domain.ErrMerchantNotFound,
		},
		{
			name: "missing required fields",
			input: CreateBranchInput{
				MerchantID: 1,
			},
			merchantRepo: &mockMerchantRepo{merchants: map[int64]*domain.Merchant{1: {ID: 1}}},
			branchRepo:   &mockBranchRepo{},
			staffRepo:    &mockStaffRepo{},
			wantErr:      domain.ErrInvalidBranch,
		},
		{
			name: "branch code already exists",
			input: CreateBranchInput{
				MerchantID: 1,
				Name:       "Branch C",
				Code:       "BR-C",
			},
			merchantRepo: &mockMerchantRepo{merchants: map[int64]*domain.Merchant{1: {ID: 1}}},
			branchRepo:   &mockBranchRepo{err: domain.ErrBranchExists},
			staffRepo:    &mockStaffRepo{},
			wantErr:      domain.ErrBranchExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := NewMerchantUsecase(tt.merchantRepo, tt.branchRepo, tt.staffRepo)
			result, err := uc.CreateBranch(context.Background(), tt.input)

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
		input        AssignStaffInput
		merchantRepo repository.MerchantRepository
		branchRepo   repository.BranchRepository
		staffRepo    repository.StaffRepository
		wantErr      error
	}{
		{
			name: "successful assignment",
			input: AssignStaffInput{
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
			input: AssignStaffInput{
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
			input: AssignStaffInput{
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
			input: AssignStaffInput{
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
			input: AssignStaffInput{
				MerchantID: 1,
				BranchID:   1,
			},
			merchantRepo: &mockMerchantRepo{merchants: map[int64]*domain.Merchant{1: {ID: 1}}},
			branchRepo:   &mockBranchRepo{branches: map[int64]*domain.Branch{1: {ID: 1}}},
			staffRepo:    &mockStaffRepo{},
			wantErr:      domain.ErrInvalidStaff,
		},
		{
			name: "concurrent duplicate assignment",
			input: AssignStaffInput{
				MerchantID: 1,
				BranchID:   1,
				UserID:     1,
				Role:       domain.StaffRoleCashier,
			},
			merchantRepo: &mockMerchantRepo{merchants: map[int64]*domain.Merchant{1: {ID: 1}}},
			branchRepo:   &mockBranchRepo{branches: map[int64]*domain.Branch{1: {ID: 1, MerchantID: 1}}},
			staffRepo: &mockStaffRepo{
				staff: map[int64]*domain.StaffMember{}, // empty so GetByUserAndBranch returns nil
				err:   domain.ErrStaffExists,           // but Create returns ErrStaffExists (unique violation)
			},
			wantErr: domain.ErrStaffExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := NewMerchantUsecase(tt.merchantRepo, tt.branchRepo, tt.staffRepo)
			result, err := uc.AssignStaff(context.Background(), tt.input)

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
			if result.UserID != tt.input.UserID {
				t.Errorf("expected user ID %d, got %d", tt.input.UserID, result.UserID)
			}
			if result.Role != tt.input.Role {
				t.Errorf("expected role %s, got %s", tt.input.Role, result.Role)
			}
		})
	}
}

func TestMerchantUsecase_UpdateStaff(t *testing.T) {
	t.Parallel()

	uc := NewMerchantUsecase(&mockMerchantRepo{}, &mockBranchRepo{}, &mockStaffRepo{
		staff: map[int64]*domain.StaffMember{
			1: {ID: 1, Role: domain.StaffRoleCashier, Status: domain.StaffStatusActive},
		},
	})

	role := domain.StaffRoleSupervisor
	status := domain.StaffStatusInactive
	result, err := uc.UpdateStaff(context.Background(), UpdateStaffInput{
		ID:     1,
		Role:   &role,
		Status: &status,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Role != domain.StaffRoleSupervisor {
		t.Errorf("expected role supervisor, got %s", result.Role)
	}
	if result.Status != domain.StaffStatusInactive {
		t.Errorf("expected status inactive, got %s", result.Status)
	}
}

func TestMerchantUsecase_RemoveStaff(t *testing.T) {
	t.Parallel()

	staffRepo := &mockStaffRepo{
		staff: map[int64]*domain.StaffMember{
			1: {ID: 1},
		},
	}
	uc := NewMerchantUsecase(&mockMerchantRepo{}, &mockBranchRepo{}, staffRepo)

	err := uc.RemoveStaff(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = uc.GetStaff(context.Background(), 1)
	if err != domain.ErrStaffNotFound {
		t.Errorf("expected ErrStaffNotFound after delete, got %v", err)
	}
}

func TestMerchantUsecase_ListStaffByBranch(t *testing.T) {
	t.Parallel()

	uc := NewMerchantUsecase(&mockMerchantRepo{}, &mockBranchRepo{}, &mockStaffRepo{
		staff: map[int64]*domain.StaffMember{
			1: {ID: 1, BranchID: 1},
			2: {ID: 2, BranchID: 1},
			3: {ID: 3, BranchID: 2},
		},
	})

	staff, err := uc.ListStaffByBranch(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(staff) != 2 {
		t.Errorf("expected 2 staff, got %d", len(staff))
	}
}

func TestMerchantUsecase_GetMyStaffAssignments(t *testing.T) {
	t.Parallel()

	uc := NewMerchantUsecase(&mockMerchantRepo{}, &mockBranchRepo{}, &mockStaffRepo{
		staff: map[int64]*domain.StaffMember{
			1: {ID: 1, UserID: 1, MerchantID: 1, BranchID: 1},
			2: {ID: 2, UserID: 1, MerchantID: 1, BranchID: 2},
			3: {ID: 3, UserID: 2, MerchantID: 1, BranchID: 1},
		},
	})

	assignments, err := uc.GetMyStaffAssignments(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(assignments) != 2 {
		t.Errorf("expected 2 assignments, got %d", len(assignments))
	}
}

func TestMerchantUsecase_SetMyDefaultBranch(t *testing.T) {
	t.Parallel()

	uc := NewMerchantUsecase(
		&mockMerchantRepo{},
		&mockBranchRepo{branches: map[int64]*domain.Branch{1: {ID: 1}}},
		&mockStaffRepo{
			staff: map[int64]*domain.StaffMember{
				1: {ID: 1, UserID: 1, BranchID: 1, MerchantID: 1},
				2: {ID: 2, UserID: 1, BranchID: 2, MerchantID: 1},
			},
		},
	)

	err := uc.SetMyDefaultBranch(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assignments, _ := uc.GetMyStaffAssignments(context.Background(), 1, 1)
	for _, a := range assignments {
		if a.BranchID == 1 && !a.IsDefault {
			t.Error("expected branch 1 to be default")
		}
		if a.BranchID == 2 && a.IsDefault {
			t.Error("expected branch 2 to not be default")
		}
	}
}

func TestMerchantUsecase_SetMyDefaultBranch_BranchNotFound(t *testing.T) {
	t.Parallel()

	uc := NewMerchantUsecase(
		&mockMerchantRepo{},
		&mockBranchRepo{},
		&mockStaffRepo{},
	)

	err := uc.SetMyDefaultBranch(context.Background(), 1, 999)
	if err != domain.ErrBranchNotFound {
		t.Errorf("expected ErrBranchNotFound, got %v", err)
	}
}
