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
	getStaff            func(context.Context, string) (*domain.StaffMember, error)
	listStaffByBranch   func(context.Context, string) ([]domain.StaffMember, error)
	listStaffByMerchant func(context.Context, string) ([]domain.StaffMember, error)
	updateStaff         func(context.Context, usecase.UpdateStaffInput) (*domain.StaffMember, error)
	removeStaff         func(context.Context, string) error
	getMyStaff          func(context.Context, string, string) ([]domain.StaffMember, error)
	setDefaultBranch    func(context.Context, string, string) error
}

func (m *mockStaffSvc) AssignStaff(ctx context.Context, i usecase.AssignStaffInput) (*domain.StaffMember, error) {
	return m.assignStaff(ctx, i)
}
func (m *mockStaffSvc) GetStaff(ctx context.Context, id string) (*domain.StaffMember, error) {
	return m.getStaff(ctx, id)
}
func (m *mockStaffSvc) ListStaffByBranch(ctx context.Context, id string) ([]domain.StaffMember, error) {
	return m.listStaffByBranch(ctx, id)
}
func (m *mockStaffSvc) ListStaffByMerchant(ctx context.Context, id string) ([]domain.StaffMember, error) {
	return m.listStaffByMerchant(ctx, id)
}
func (m *mockStaffSvc) UpdateStaff(ctx context.Context, i usecase.UpdateStaffInput) (*domain.StaffMember, error) {
	return m.updateStaff(ctx, i)
}
func (m *mockStaffSvc) RemoveStaff(ctx context.Context, id string) error {
	return m.removeStaff(ctx, id)
}
func (m *mockStaffSvc) GetMyStaffAssignments(ctx context.Context, u, mID string) ([]domain.StaffMember, error) {
	return m.getMyStaff(ctx, u, mID)
}
func (m *mockStaffSvc) SetMyDefaultBranch(ctx context.Context, u, b string) error {
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
			body: `{"branch_id":"b1","user_id":"u1","role":"cashier","is_default":true}`,
			mock: &mockStaffSvc{
				assignStaff: func(_ context.Context, i usecase.AssignStaffInput) (*domain.StaffMember, error) {
					return &domain.StaffMember{
						ID: "s1", MerchantID: i.MerchantID, BranchID: i.BranchID,
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
			body:       `{"branch_id":"b1"}`,
			mock:       &mockStaffSvc{},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "conflict",
			body: `{"branch_id":"b1","user_id":"u1","role":"cashier"}`,
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
			c.Set("merchant_id", "m1")

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
			id:   "s1",
			mock: &mockStaffSvc{
				getStaff: func(_ context.Context, id string) (*domain.StaffMember, error) {
					return &domain.StaffMember{ID: id, UserID: "u1", Role: domain.StaffRoleCashier}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			id:   "nonexistent",
			mock: &mockStaffSvc{
				getStaff: func(_ context.Context, _ string) (*domain.StaffMember, error) {
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
		listStaffByBranch: func(_ context.Context, branchID string) ([]domain.StaffMember, error) {
			return []domain.StaffMember{
				{ID: "s1", BranchID: branchID, Role: domain.StaffRoleCashier},
				{ID: "s2", BranchID: branchID, Role: domain.StaffRoleManager},
			}, nil
		},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/branches/b1/staff", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("branchId")
	c.SetParamValues("b1")

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
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/staff/s1", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("s1")

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
		removeStaff: func(_ context.Context, id string) error {
			return nil
		},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/staff/s1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("s1")

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
		getMyStaff: func(_ context.Context, userID, merchantID string) ([]domain.StaffMember, error) {
			return []domain.StaffMember{
				{ID: "s1", UserID: userID, MerchantID: merchantID, BranchID: "b1"},
			}, nil
		},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/assignments", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("merchant_id", "m1")
	c.Set("user_id", "u1")

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
			body: `{"branch_id":"b1"}`,
			mock: &mockStaffSvc{
				setDefaultBranch: func(_ context.Context, _, _ string) error { return nil },
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
			body:       `{"branch_id":""}`,
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
			c.Set("user_id", "u1")

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
