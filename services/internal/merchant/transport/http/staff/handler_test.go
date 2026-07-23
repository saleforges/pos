package staff

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/internal/merchant/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
	"github.com/saleforge/pos/services/pkg/pagination"
)

type mockStaffSvc struct {
	assignFn    func(context.Context, usecase.AssignStaffParams) (*domain.StaffMember, error)
	getFn       func(context.Context, int64) (*domain.StaffMember, error)
	listBranchFn func(context.Context, int64, pagination.Params) ([]domain.StaffMember, *pagination.Metadata, error)
	listMerchantFn func(context.Context, int64, pagination.Params) ([]domain.StaffMember, *pagination.Metadata, error)
	getAssignFn func(context.Context, int64, int64) ([]domain.StaffMember, error)
	setDefaultFn func(context.Context, int64, int64) error
	updateFn    func(context.Context, usecase.UpdateStaffParams) (*domain.StaffMember, error)
	removeFn    func(context.Context, int64) error
}

func (m *mockStaffSvc) AssignStaff(ctx context.Context, p usecase.AssignStaffParams) (*domain.StaffMember, error) {
	return m.assignFn(ctx, p)
}
func (m *mockStaffSvc) GetStaff(ctx context.Context, id int64) (*domain.StaffMember, error) {
	return m.getFn(ctx, id)
}
func (m *mockStaffSvc) ListStaffByBranch(ctx context.Context, branchID int64, p pagination.Params) ([]domain.StaffMember, *pagination.Metadata, error) {
	return m.listBranchFn(ctx, branchID, p)
}
func (m *mockStaffSvc) ListStaffByMerchant(ctx context.Context, merchantID int64, p pagination.Params) ([]domain.StaffMember, *pagination.Metadata, error) {
	return m.listMerchantFn(ctx, merchantID, p)
}
func (m *mockStaffSvc) GetMyStaffAssignments(ctx context.Context, userID, merchantID int64) ([]domain.StaffMember, error) {
	return m.getAssignFn(ctx, userID, merchantID)
}
func (m *mockStaffSvc) SetMyDefaultBranch(ctx context.Context, userID, branchID int64) error {
	return m.setDefaultFn(ctx, userID, branchID)
}
func (m *mockStaffSvc) UpdateStaff(ctx context.Context, p usecase.UpdateStaffParams) (*domain.StaffMember, error) {
	return m.updateFn(ctx, p)
}
func (m *mockStaffSvc) RemoveStaff(ctx context.Context, id int64) error {
	return m.removeFn(ctx, id)
}

func TestStaffHandler_AssignStaff(t *testing.T) {
	t.Parallel()

	e := echo.New()
	body := `{"branchId":1,"userId":1,"role":"cashier"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/staff", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(httputil.ContextKeyMerchantID, int64(1))

	h := NewHandler(&mockStaffSvc{
		assignFn: func(_ context.Context, p usecase.AssignStaffParams) (*domain.StaffMember, error) {
			return &domain.StaffMember{ID: 1, UserID: p.UserID, Role: p.Role}, nil
		},
	})

	if err := h.AssignStaff(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestStaffHandler_GetStaff(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	h := NewHandler(&mockStaffSvc{
		getFn: func(_ context.Context, id int64) (*domain.StaffMember, error) {
			return &domain.StaffMember{ID: id, Role: domain.StaffRoleCashier}, nil
		},
	})

	if err := h.GetStaff(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestStaffHandler_ListStaffByMerchant(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/staff?offset=0&limit=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(httputil.ContextKeyMerchantID, int64(1))

	h := NewHandler(&mockStaffSvc{
		listMerchantFn: func(_ context.Context, merchantID int64, p pagination.Params) ([]domain.StaffMember, *pagination.Metadata, error) {
			return []domain.StaffMember{{ID: 1, Role: domain.StaffRoleCashier}}, &pagination.Metadata{Total: 1, Offset: 0, Limit: 10, Count: 1}, nil
		},
	})

	if err := h.ListStaffByMerchant(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestStaffHandler_RemoveStaff(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	h := NewHandler(&mockStaffSvc{
		removeFn: func(_ context.Context, id int64) error {
			return nil
		},
	})

	if err := h.RemoveStaff(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
