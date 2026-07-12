package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/internal/merchant/usecase"
)

type mockStaffSvc struct {
	assignStaff         func(context.Context, usecase.AssignStaffInput) (*domain.StaffMember, error)
	getStaff            func(context.Context, int64) (*domain.StaffMember, error)
	listStaffByBranch   func(context.Context, int64) ([]domain.StaffMember, error)
	listStaffByMerchant func(context.Context, int64) ([]domain.StaffMember, error)
	updateStaff         func(context.Context, usecase.UpdateStaffInput) (*domain.StaffMember, error)
	removeStaff         func(context.Context, int64) error
	getMyStaff          func(context.Context, int64, int64) ([]domain.StaffMember, error)
	setDefaultBranch    func(context.Context, int64, int64) error
}

func (m *mockStaffSvc) AssignStaff(ctx context.Context, i usecase.AssignStaffInput) (*domain.StaffMember, error) {
	return m.assignStaff(ctx, i)
}
func (m *mockStaffSvc) GetStaff(ctx context.Context, id int64) (*domain.StaffMember, error) {
	return m.getStaff(ctx, id)
}
func (m *mockStaffSvc) ListStaffByBranch(ctx context.Context, id int64) ([]domain.StaffMember, error) {
	return m.listStaffByBranch(ctx, id)
}
func (m *mockStaffSvc) ListStaffByMerchant(ctx context.Context, id int64) ([]domain.StaffMember, error) {
	return m.listStaffByMerchant(ctx, id)
}
func (m *mockStaffSvc) UpdateStaff(ctx context.Context, i usecase.UpdateStaffInput) (*domain.StaffMember, error) {
	return m.updateStaff(ctx, i)
}
func (m *mockStaffSvc) RemoveStaff(ctx context.Context, id int64) error {
	return m.removeStaff(ctx, id)
}
func (m *mockStaffSvc) GetMyStaffAssignments(ctx context.Context, u, mID int64) ([]domain.StaffMember, error) {
	return m.getMyStaff(ctx, u, mID)
}
func (m *mockStaffSvc) SetMyDefaultBranch(ctx context.Context, u, b int64) error {
	return m.setDefaultBranch(ctx, u, b)
}

func TestStaffHandler_AssignStaff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		mock       *mockStaffSvc
		wantStatus int
	}{
		{
			name: "success",
			body: `{"branch_id":1,"user_id":1,"role":"cashier","is_default":true}`,
			mock: &mockStaffSvc{
				assignStaff: func(_ context.Context, i usecase.AssignStaffInput) (*domain.StaffMember, error) {
					return &domain.StaffMember{
						ID: 1, MerchantID: i.MerchantID, BranchID: i.BranchID,
						UserID: i.UserID, Role: i.Role, IsDefault: i.IsDefault,
						Status: domain.StaffStatusActive,
					}, nil
				},
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid json",
			body:       `{bad}`,
			mock:       &mockStaffSvc{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing fields",
			body:       `{"branch_id":1}`,
			mock:       &mockStaffSvc{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "conflict",
			body: `{"branch_id":1,"user_id":1,"role":"cashier"}`,
			mock: &mockStaffSvc{
				assignStaff: func(_ context.Context, _ usecase.AssignStaffInput) (*domain.StaffMember, error) {
					return nil, domain.ErrStaffExists
				},
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/staff", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("merchant_id", int64(1))

			h := NewStaffHandler(tt.mock)
			err := h.AssignStaff(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestStaffHandler_GetStaff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		mock       *mockStaffSvc
		wantStatus int
	}{
		{
			name: "success",
			id:   "1",
			mock: &mockStaffSvc{
				getStaff: func(_ context.Context, id int64) (*domain.StaffMember, error) {
					return &domain.StaffMember{ID: id, UserID: 1, Role: domain.StaffRoleCashier}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			id:   "999",
			mock: &mockStaffSvc{
				getStaff: func(_ context.Context, _ int64) (*domain.StaffMember, error) {
					return nil, domain.ErrStaffNotFound
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/staff/"+tt.id, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			h := NewStaffHandler(tt.mock)
			err := h.GetStaff(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestStaffHandler_ListStaffByBranch(t *testing.T) {
	t.Parallel()

	mock := &mockStaffSvc{
		listStaffByBranch: func(_ context.Context, branchID int64) ([]domain.StaffMember, error) {
			return []domain.StaffMember{
				{ID: 1, BranchID: branchID, Role: domain.StaffRoleCashier},
				{ID: 2, BranchID: branchID, Role: domain.StaffRoleManager},
			}, nil
		},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/branches/1/staff", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("branchId")
	c.SetParamValues("1")

	h := NewStaffHandler(mock)
	err := h.ListStaffByBranch(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestStaffHandler_UpdateStaff(t *testing.T) {
	t.Parallel()

	mock := &mockStaffSvc{
		updateStaff: func(_ context.Context, i usecase.UpdateStaffInput) (*domain.StaffMember, error) {
			return &domain.StaffMember{ID: i.ID, Role: *i.Role}, nil
		},
	}

	e := echo.New()
	body := `{"role":"supervisor"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/staff/1", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	h := NewStaffHandler(mock)
	err := h.UpdateStaff(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestStaffHandler_RemoveStaff(t *testing.T) {
	t.Parallel()

	mock := &mockStaffSvc{
		removeStaff: func(_ context.Context, id int64) error {
			return nil
		},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/staff/1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	h := NewStaffHandler(mock)
	err := h.RemoveStaff(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestStaffHandler_MyStaffAssignments(t *testing.T) {
	t.Parallel()

	mock := &mockStaffSvc{
		getMyStaff: func(_ context.Context, userID, merchantID int64) ([]domain.StaffMember, error) {
			return []domain.StaffMember{
				{ID: 1, UserID: userID, MerchantID: merchantID, BranchID: 1},
			}, nil
		},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/assignments", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("merchant_id", int64(1))
	c.Set("user_id", int64(1))

	h := NewStaffHandler(mock)
	err := h.MyStaffAssignments(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestStaffHandler_SetMyDefaultBranch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		mock       *mockStaffSvc
		wantStatus int
	}{
		{
			name: "success",
			body: `{"branch_id":1}`,
			mock: &mockStaffSvc{
				setDefaultBranch: func(_ context.Context, _, _ int64) error { return nil },
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid json",
			body:       `{bad}`,
			mock:       &mockStaffSvc{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty branch id",
			body:       `{}`,
			mock:       &mockStaffSvc{},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodPut, "/api/v1/me/default-branch", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("user_id", int64(1))

			h := NewStaffHandler(tt.mock)
			err := h.SetMyDefaultBranch(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}
