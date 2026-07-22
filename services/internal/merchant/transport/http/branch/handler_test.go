package branch

import (
	"context"
	"encoding/json"
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

type mockBranchSvc struct {
	createFn func(context.Context, usecase.CreateBranchParams) (*domain.Branch, error)
	getFn    func(context.Context, int64) (*domain.Branch, error)
	listFn   func(context.Context, int64, pagination.Params) ([]domain.Branch, *pagination.Metadata, error)
	updateFn func(context.Context, usecase.UpdateBranchParams) (*domain.Branch, error)
	deleteFn func(context.Context, int64) error
}

func (m *mockBranchSvc) CreateBranch(ctx context.Context, p usecase.CreateBranchParams) (*domain.Branch, error) {
	return m.createFn(ctx, p)
}
func (m *mockBranchSvc) GetBranch(ctx context.Context, id int64) (*domain.Branch, error) {
	return m.getFn(ctx, id)
}
func (m *mockBranchSvc) ListBranches(ctx context.Context, merchantID int64, p pagination.Params) ([]domain.Branch, *pagination.Metadata, error) {
	return m.listFn(ctx, merchantID, p)
}
func (m *mockBranchSvc) UpdateBranch(ctx context.Context, p usecase.UpdateBranchParams) (*domain.Branch, error) {
	return m.updateFn(ctx, p)
}
func (m *mockBranchSvc) DeleteBranch(ctx context.Context, id int64) error {
	return m.deleteFn(ctx, id)
}

func TestBranchHandler_CreateBranch(t *testing.T) {
	t.Parallel()

	e := echo.New()
	body := `{"name":"Branch A","code":"BR-A"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/branches", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(httputil.ContextKeyMerchantID, int64(1))

	h := NewHandler(&mockBranchSvc{
		createFn: func(_ context.Context, p usecase.CreateBranchParams) (*domain.Branch, error) {
			return &domain.Branch{ID: 1, Name: p.Name, Code: p.Code}, nil
		},
	})

	if err := h.CreateBranch(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestBranchHandler_GetBranch(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	h := NewHandler(&mockBranchSvc{
		getFn: func(_ context.Context, id int64) (*domain.Branch, error) {
			return &domain.Branch{ID: id, Name: "Branch A"}, nil
		},
	})

	if err := h.GetBranch(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["message"] != "success" {
		t.Errorf("expected success message, got %v", resp["message"])
	}
}

func TestBranchHandler_ListBranches(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/branches?offset=0&limit=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(httputil.ContextKeyMerchantID, int64(1))

	h := NewHandler(&mockBranchSvc{
		listFn: func(_ context.Context, merchantID int64, p pagination.Params) ([]domain.Branch, *pagination.Metadata, error) {
			return []domain.Branch{{ID: 1, Name: "B1"}}, &pagination.Metadata{Total: 1, Offset: 0, Limit: 10, ReturnCount: 1}, nil
		},
	})

	if err := h.ListBranches(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["pagination"] == nil {
		t.Error("expected pagination metadata")
	}
}
