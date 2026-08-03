package stock

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/inventory/domain"
	"github.com/saleforge/pos/services/internal/inventory/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
)

type mockStockSvc struct {
	createFn func(context.Context, usecase.CreateStockParams) (*domain.Stock, error)
	getFn    func(context.Context, int64, int64) (*domain.Stock, error)
	listFn   func(context.Context, int64) ([]domain.Stock, error)
	updateFn func(context.Context, usecase.UpdateStockParams) (*domain.Stock, error)
}

func (m *mockStockSvc) Create(ctx context.Context, p usecase.CreateStockParams) (*domain.Stock, error) {
	if m.createFn != nil {
		return m.createFn(ctx, p)
	}
	return &domain.Stock{ID: 1, MerchantID: p.MerchantID, BranchID: p.BranchID, ProductItemID: p.ProductItemID, Available: p.Available}, nil
}

func (m *mockStockSvc) GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Stock, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id, merchantID)
	}
	return &domain.Stock{ID: id, MerchantID: merchantID, BranchID: 1, ProductItemID: 1, Available: 100}, nil
}

func (m *mockStockSvc) List(ctx context.Context, merchantID int64) ([]domain.Stock, error) {
	if m.listFn != nil {
		return m.listFn(ctx, merchantID)
	}
	return []domain.Stock{{ID: 1, MerchantID: merchantID, BranchID: 1, ProductItemID: 1, Available: 100}}, nil
}

func (m *mockStockSvc) Update(ctx context.Context, p usecase.UpdateStockParams) (*domain.Stock, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, p)
	}
	return &domain.Stock{ID: p.ID, MerchantID: p.MerchantID, BranchID: 1, ProductItemID: 1, Available: p.Available}, nil
}

func (m *mockStockSvc) Deduct(context.Context, usecase.AdjustStockParams) error { return nil }

func (m *mockStockSvc) Restore(context.Context, usecase.AdjustStockParams) error { return nil }

func withMerchant(c echo.Context) echo.Context {
	c.Set(httputil.ContextKeyMerchantID, int64(1))
	return c
}

func TestCreate(t *testing.T) {
	t.Parallel()

	t.Run("valid request returns 201", func(t *testing.T) {
		e := echo.New()
		body := `{"branchId":1,"productItemId":1,"available":100}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c = withMerchant(c)

		h := NewHandler(&mockStockSvc{})
		if err := h.Create(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rec.Code)
		}

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp["data"] == nil {
			t.Error("expected data in response")
		}
	})

	t.Run("empty body returns 400", func(t *testing.T) {
		e := echo.New()
		body := `{}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c = withMerchant(c)

		h := NewHandler(&mockStockSvc{
			createFn: func(_ context.Context, p usecase.CreateStockParams) (*domain.Stock, error) {
				if p.BranchID == 0 || p.ProductItemID == 0 {
					return nil, domain.ErrInvalidStock
				}
				return &domain.Stock{ID: 1, MerchantID: p.MerchantID, BranchID: p.BranchID, ProductItemID: p.ProductItemID, Available: p.Available}, nil
			},
		})
		if err := h.Create(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})
}

func TestGetByID(t *testing.T) {
	t.Parallel()

	t.Run("returns 200", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")
		c = withMerchant(c)

		h := NewHandler(&mockStockSvc{})
		if err := h.GetByID(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("not found returns 404", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("999")
		c = withMerchant(c)

		h := NewHandler(&mockStockSvc{
			getFn: func(_ context.Context, id int64, merchantID int64) (*domain.Stock, error) {
				return nil, domain.ErrStockNotFound
			},
		})
		if err := h.GetByID(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})
}

func TestList(t *testing.T) {
	t.Parallel()

	t.Run("returns 200 with items", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c = withMerchant(c)

		h := NewHandler(&mockStockSvc{})
		if err := h.List(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	t.Run("returns 200", func(t *testing.T) {
		e := echo.New()
		body := `{"available":200}`
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")
		c = withMerchant(c)

		h := NewHandler(&mockStockSvc{})
		if err := h.Update(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}
