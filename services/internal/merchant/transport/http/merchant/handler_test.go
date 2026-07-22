package merchant

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
	"github.com/saleforge/pos/services/pkg/pagination"
)

type mockMerchantSvc struct {
	createFn func(context.Context, usecase.CreateMerchantParams) (*domain.Merchant, error)
	getFn    func(context.Context, int64) (*domain.Merchant, error)
	listFn   func(context.Context, pagination.Params) ([]domain.Merchant, *pagination.Metadata, error)
	updateFn func(context.Context, usecase.UpdateMerchantParams) (*domain.Merchant, error)
	deleteFn func(context.Context, int64) error
}

func (m *mockMerchantSvc) CreateMerchant(ctx context.Context, p usecase.CreateMerchantParams) (*domain.Merchant, error) {
	return m.createFn(ctx, p)
}
func (m *mockMerchantSvc) GetMerchant(ctx context.Context, id int64) (*domain.Merchant, error) {
	return m.getFn(ctx, id)
}
func (m *mockMerchantSvc) ListMerchants(ctx context.Context, p pagination.Params) ([]domain.Merchant, *pagination.Metadata, error) {
	return m.listFn(ctx, p)
}
func (m *mockMerchantSvc) UpdateMerchant(ctx context.Context, p usecase.UpdateMerchantParams) (*domain.Merchant, error) {
	return m.updateFn(ctx, p)
}
func (m *mockMerchantSvc) DeleteMerchant(ctx context.Context, id int64) error {
	return m.deleteFn(ctx, id)
}

func TestMerchantHandler_Create(t *testing.T) {
	t.Parallel()

	e := echo.New()
	body := `{"name":"Toko Maju","email":"toko@maju.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/merchants", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewHandler(&mockMerchantSvc{
		createFn: func(_ context.Context, p usecase.CreateMerchantParams) (*domain.Merchant, error) {
			return &domain.Merchant{ID: 1, Name: p.Name, Email: p.Email}, nil
		},
	})

	if err := h.Create(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rec.Code)
	}
}

func TestMerchantHandler_Create_ValidationError(t *testing.T) {
	t.Parallel()

	e := echo.New()
	body := `{"name":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/merchants", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewHandler(&mockMerchantSvc{})

	if err := h.Create(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestMerchantHandler_Get(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	h := NewHandler(&mockMerchantSvc{
		getFn: func(_ context.Context, id int64) (*domain.Merchant, error) {
			return &domain.Merchant{ID: id, Name: "Test"}, nil
		},
	})

	if err := h.Get(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["message"] != "success" {
		t.Errorf("expected success message")
	}
}

func TestMerchantHandler_Get_NotFound(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("999")

	h := NewHandler(&mockMerchantSvc{
		getFn: func(_ context.Context, id int64) (*domain.Merchant, error) {
			return nil, domain.ErrMerchantNotFound
		},
	})

	if err := h.Get(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestMerchantHandler_List(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/merchants?offset=0&limit=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewHandler(&mockMerchantSvc{
		listFn: func(_ context.Context, p pagination.Params) ([]domain.Merchant, *pagination.Metadata, error) {
			return []domain.Merchant{{ID: 1, Name: "M1"}}, &pagination.Metadata{Total: 1, Offset: 0, Limit: 10, ReturnCount: 1}, nil
		},
	})

	if err := h.List(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["pagination"] == nil {
		t.Error("expected pagination metadata in response")
	}
}

func TestMerchantHandler_Delete(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	h := NewHandler(&mockMerchantSvc{
		deleteFn: func(_ context.Context, id int64) error {
			return nil
		},
	})

	if err := h.Delete(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
