package productitem

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
)

type mockItemSvc struct {
	createFn   func(context.Context, usecase.CreateProductItemParams) (*domain.ProductItem, error)
	getFn      func(context.Context, int64, int64) (*domain.ProductItem, error)
	listFn     func(context.Context, int64) ([]domain.ProductItem, error)
	listProdFn func(context.Context, int64, int64) ([]domain.ProductItem, error)
	updateFn   func(context.Context, usecase.UpdateProductItemParams) (*domain.ProductItem, error)
	deleteFn   func(context.Context, int64, int64) error
	restoreFn  func(context.Context, int64, int64) (*domain.ProductItem, error)
}

func (m *mockItemSvc) Create(ctx context.Context, p usecase.CreateProductItemParams) (*domain.ProductItem, error) {
	if m.createFn != nil {
		return m.createFn(ctx, p)
	}
	return &domain.ProductItem{ID: 1, ProductID: p.ProductID, MerchantID: p.MerchantID, Name: p.Name, Price: domain.Price{Amount: p.PriceAmount, Currency: "IDR"}, Status: domain.ProductItemStatusActive}, nil
}

func (m *mockItemSvc) GetByID(ctx context.Context, id int64, merchantID int64) (*domain.ProductItem, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id, merchantID)
	}
	return &domain.ProductItem{ID: id, MerchantID: merchantID, Name: "Item", UnitID: int64Ptr(1), Price: domain.Price{Amount: 10000, Currency: "IDR"}}, nil
}

func (m *mockItemSvc) ListByProduct(ctx context.Context, productID int64, merchantID int64) ([]domain.ProductItem, error) {
	if m.listProdFn != nil {
		return m.listProdFn(ctx, productID, merchantID)
	}
	return []domain.ProductItem{{ID: 1, ProductID: productID, MerchantID: merchantID, Name: "Item", UnitID: int64Ptr(1), Price: domain.Price{Amount: 10000, Currency: "IDR"}}}, nil
}

func (m *mockItemSvc) ListByMerchant(ctx context.Context, merchantID int64) ([]domain.ProductItem, error) {
	return []domain.ProductItem{{ID: 1, ProductID: 1, MerchantID: merchantID, Name: "Item", Price: domain.Price{Amount: 10000, Currency: "IDR"}}}, nil
}

func (m *mockItemSvc) Update(ctx context.Context, p usecase.UpdateProductItemParams) (*domain.ProductItem, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, p)
	}
	return &domain.ProductItem{ID: p.ID, Name: "Updated", UnitID: int64Ptr(1), Price: domain.Price{Amount: 10000, Currency: "IDR"}}, nil
}

func (m *mockItemSvc) Delete(ctx context.Context, id int64, merchantID int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id, merchantID)
	}
	return nil
}

func (m *mockItemSvc) Restore(ctx context.Context, id int64, merchantID int64) (*domain.ProductItem, error) {
	if m.restoreFn != nil {
		return m.restoreFn(ctx, id, merchantID)
	}
	return &domain.ProductItem{ID: id, MerchantID: merchantID, Name: "Restored", UnitID: int64Ptr(1), Price: domain.Price{Amount: 10000, Currency: "IDR"}}, nil
}

func int64Ptr(v int64) *int64 { return &v }

func withMerchant(c echo.Context) echo.Context {
	c.Set(httputil.ContextKeyMerchantID, int64(1))
	return c
}

func TestCreate(t *testing.T) {
	t.Parallel()

	t.Run("valid request returns 201", func(t *testing.T) {
		e := echo.New()
		body := `{"name":"Marning Curah","unitId":1,"price":15000}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("productId")
		c.SetParamValues("1")
		c = withMerchant(c)

		h := NewHandler(&mockItemSvc{})
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

	t.Run("empty name returns 400", func(t *testing.T) {
		e := echo.New()
		body := `{"name":"","unitId":1}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("productId")
		c.SetParamValues("1")
		c = withMerchant(c)

		h := NewHandler(&mockItemSvc{})
		if err := h.Create(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("empty body returns 400", func(t *testing.T) {
		e := echo.New()
		body := `{}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("productId")
		c.SetParamValues("1")
		c = withMerchant(c)

		h := NewHandler(&mockItemSvc{})
		if err := h.Create(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("valid request with nullable unit", func(t *testing.T) {
		e := echo.New()
		body := `{"name":"Es Teh","price":3000}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("productId")
		c.SetParamValues("1")
		c = withMerchant(c)

		h := NewHandler(&mockItemSvc{
			createFn: func(_ context.Context, p usecase.CreateProductItemParams) (*domain.ProductItem, error) {
				if p.UnitID != nil {
					t.Error("expected nil unitId for simple product")
				}
				return &domain.ProductItem{ID: 1, ProductID: p.ProductID, MerchantID: p.MerchantID, Name: p.Name, Price: domain.Price{Amount: p.PriceAmount, Currency: "IDR"}}, nil
			},
		})
		if err := h.Create(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rec.Code)
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

		h := NewHandler(&mockItemSvc{})
		if err := h.GetByID(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestListByProduct(t *testing.T) {
	t.Parallel()

	t.Run("returns 200 with items", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("productId")
		c.SetParamValues("1")
		c = withMerchant(c)

		h := NewHandler(&mockItemSvc{})
		if err := h.ListByProduct(c); err != nil {
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
		body := `{"name":"New Name"}`
		req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")
		c = withMerchant(c)

		h := NewHandler(&mockItemSvc{})
		if err := h.Update(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestDelete(t *testing.T) {
	t.Parallel()

	t.Run("success returns 200", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")
		c = withMerchant(c)

		h := NewHandler(&mockItemSvc{})
		if err := h.Delete(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("not found returns 404", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("999")
		c = withMerchant(c)

		h := NewHandler(&mockItemSvc{
			deleteFn: func(_ context.Context, id int64, merchantID int64) error {
				return domain.ErrProductItemNotFound
			},
		})
		if err := h.Delete(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})
}

func (m *mockItemSvc) SetBranchPrice(ctx context.Context, _, _, _ int64, _ float64, _ string) error {
	return nil
}
func (m *mockItemSvc) DeleteBranchPrice(ctx context.Context, _, _, _ int64) error { return nil }
func (m *mockItemSvc) GetBranchPrice(ctx context.Context, _, _, _ int64) (*domain.Price, error) {
	return nil, nil
}
