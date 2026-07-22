package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/merchant/domain"
)

type mockMerchantRepo struct {
	merchants map[int64]*domain.Merchant
	err       error
}

func (m *mockMerchantRepo) Create(_ context.Context, merchant *domain.Merchant) error {
	if m.err != nil { return m.err }
	if m.merchants == nil { m.merchants = make(map[int64]*domain.Merchant) }
	merchant.ID = int64(len(m.merchants) + 1)
	m.merchants[merchant.ID] = merchant
	return nil
}

func (m *mockMerchantRepo) GetByID(_ context.Context, id int64) (*domain.Merchant, error) {
	if m.err != nil { return nil, m.err }
	merchant, ok := m.merchants[id]
	if !ok { return nil, domain.ErrMerchantNotFound }
	return merchant, nil
}

func (m *mockMerchantRepo) List(_ context.Context, offset, limit int) ([]domain.Merchant, int64, error) {
	if m.err != nil { return nil, 0, m.err }
	var result []domain.Merchant
	for _, merchant := range m.merchants {
		result = append(result, *merchant)
	}
	total := int64(len(result))
	if offset >= len(result) {
		return []domain.Merchant{}, total, nil
	}
	end := offset + limit
	if limit == -1 { end = len(result); offset = 0 }
	if end > len(result) { end = len(result) }
	return result[offset:end], total, nil
}

func (m *mockMerchantRepo) Update(_ context.Context, merchant *domain.Merchant) error {
	if m.err != nil { return m.err }
	m.merchants[merchant.ID] = merchant
	return nil
}

func (m *mockMerchantRepo) Delete(_ context.Context, id int64) error {
	if m.err != nil { return m.err }
	delete(m.merchants, id)
	return nil
}

type mockBranchRepo struct {
	branches map[int64]*domain.Branch
	err      error
}

func (m *mockBranchRepo) Create(_ context.Context, branch *domain.Branch) error {
	if m.err != nil { return m.err }
	if m.branches == nil { m.branches = make(map[int64]*domain.Branch) }
	branch.ID = int64(len(m.branches) + 1)
	m.branches[branch.ID] = branch
	return nil
}

func (m *mockBranchRepo) GetByID(_ context.Context, id int64) (*domain.Branch, error) {
	if m.err != nil { return nil, m.err }
	branch, ok := m.branches[id]
	if !ok { return nil, domain.ErrBranchNotFound }
	return branch, nil
}

func (m *mockBranchRepo) ListByMerchant(_ context.Context, merchantID int64, offset, limit int) ([]domain.Branch, int64, error) {
	if m.err != nil { return nil, 0, m.err }
	var result []domain.Branch
	for _, branch := range m.branches {
		if branch.MerchantID == merchantID {
			result = append(result, *branch)
		}
	}
	total := int64(len(result))
	if offset >= len(result) {
		return []domain.Branch{}, total, nil
	}
	end := offset + limit
	if limit == -1 { end = len(result); offset = 0 }
	if end > len(result) { end = len(result) }
	return result[offset:end], total, nil
}

func (m *mockBranchRepo) Update(_ context.Context, branch *domain.Branch) error {
	if m.err != nil { return m.err }
	m.branches[branch.ID] = branch
	return nil
}

func (m *mockBranchRepo) Delete(_ context.Context, id int64) error {
	if m.err != nil { return m.err }
	delete(m.branches, id)
	return nil
}

type mockStaffRepo struct {
	staff     map[int64]*domain.StaffMember
	err       error
	createErr error
}

func (m *mockStaffRepo) Create(_ context.Context, s *domain.StaffMember) error {
	if m.createErr != nil { return m.createErr }
	if m.err != nil { return m.err }
	if m.staff == nil { m.staff = make(map[int64]*domain.StaffMember) }
	s.ID = int64(len(m.staff) + 1)
	m.staff[s.ID] = s
	return nil
}

func (m *mockStaffRepo) GetByID(_ context.Context, id int64) (*domain.StaffMember, error) {
	if m.err != nil { return nil, m.err }
	s, ok := m.staff[id]
	if !ok { return nil, domain.ErrStaffNotFound }
	return s, nil
}

func (m *mockStaffRepo) ListByBranch(_ context.Context, branchID int64, offset, limit int) ([]domain.StaffMember, int64, error) {
	if m.err != nil { return nil, 0, m.err }
	var result []domain.StaffMember
	for _, s := range m.staff {
		if s.BranchID == branchID { result = append(result, *s) }
	}
	total := int64(len(result))
	if offset >= len(result) {
		return []domain.StaffMember{}, total, nil
	}
	end := offset + limit
	if limit == -1 { end = len(result); offset = 0 }
	if end > len(result) { end = len(result) }
	return result[offset:end], total, nil
}

func (m *mockStaffRepo) ListByMerchant(_ context.Context, merchantID int64, offset, limit int) ([]domain.StaffMember, int64, error) {
	if m.err != nil { return nil, 0, m.err }
	var result []domain.StaffMember
	for _, s := range m.staff {
		if s.MerchantID == merchantID { result = append(result, *s) }
	}
	total := int64(len(result))
	if offset >= len(result) {
		return []domain.StaffMember{}, total, nil
	}
	end := offset + limit
	if limit == -1 { end = len(result); offset = 0 }
	if end > len(result) { end = len(result) }
	return result[offset:end], total, nil
}

func (m *mockStaffRepo) GetByUserAndBranch(_ context.Context, userID, branchID int64) (*domain.StaffMember, error) {
	if m.err != nil { return nil, m.err }
	for _, s := range m.staff {
		if s.UserID == userID && s.BranchID == branchID { return s, nil }
	}
	return nil, domain.ErrStaffNotFound
}

func (m *mockStaffRepo) ListByUserAndMerchant(_ context.Context, userID, merchantID int64) ([]domain.StaffMember, error) {
	if m.err != nil { return nil, m.err }
	var result []domain.StaffMember
	for _, s := range m.staff {
		if s.UserID == userID && s.MerchantID == merchantID { result = append(result, *s) }
	}
	if result == nil { return []domain.StaffMember{}, nil }
	return result, nil
}

func (m *mockStaffRepo) SetDefaultBranch(_ context.Context, userID, branchID int64) error {
	if m.err != nil { return m.err }
	for _, s := range m.staff {
		if s.UserID == userID { s.IsDefault = s.BranchID == branchID }
	}
	return nil
}

func (m *mockStaffRepo) Update(_ context.Context, s *domain.StaffMember) error {
	if m.err != nil { return m.err }
	m.staff[s.ID] = s
	return nil
}

func (m *mockStaffRepo) Delete(_ context.Context, id int64) error {
	if m.err != nil { return m.err }
	delete(m.staff, id)
	return nil
}
