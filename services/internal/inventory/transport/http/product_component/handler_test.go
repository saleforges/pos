package productcomponent

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

type mockComponentSvc struct {
	createFn func(context.Context, usecase.CreateProductComponentParams) (*domain.ProductComponent, error)
	getFn    func(context.Context, int64, int64) (*domain.ProductComponent, error)
	listFn   func(context.Context, int64) ([]domain.ProductComponent, error)
	updateFn func(context.Context, usecase.UpdateProductComponentParams) (*domain.ProductComponent, error)
	deleteFn func(context.Context, int64, int64) error
}

func (m *mockComponentSvc) Create(ctx context.Context, p usecase.CreateProductComponentParams) (*domain.ProductComponent, error) {
	if m.createFn != nil {
		return m.createFn(ctx, p)
	}
	items := make([]domain.ProductComponentItem, len(p.Items))
	for i, item := range p.Items {
		items[i] = domain.ProductComponentItem{
			ID: int64(i + 1), ProductComponentID: 1,
			ComponentProductItemID: item.ComponentProductItemID,
			Quantity:               item.Quantity,
			UnitID:                 item.UnitID,
		}
	}
	return &domain.ProductComponent{ID: 1, MerchantID: p.MerchantID, ProductItemID: p.ProductItemID, Items: items}, nil
}

func (m *mockComponentSvc) GetByProductItem(ctx context.Context, productItemID int64, merchantID int64) (*domain.ProductComponent, error) {
	if m.getFn != nil {
		return m.getFn(ctx, productItemID, merchantID)
	}
	return &domain.ProductComponent{
		ID: 1, MerchantID: merchantID, ProductItemID: productItemID,
		Items: []domain.ProductComponentItem{{ID: 1, ProductComponentID: 1, ComponentProductItemID: 2, Quantity: 2, UnitID: 1}},
	}, nil
}

func (m *mockComponentSvc) List(ctx context.Context, merchantID int64) ([]domain.ProductComponent, error) {
	if m.listFn != nil {
		return m.listFn(ctx, merchantID)
	}
	return []domain.ProductComponent{{
		ID: 1, MerchantID: merchantID, ProductItemID: 1,
		Items: []domain.ProductComponentItem{{ID: 1, ProductComponentID: 1, ComponentProductItemID: 2, Quantity: 2, UnitID: 1}},
	}}, nil
}

func (m *mockComponentSvc) Update(ctx context.Context, p usecase.UpdateProductComponentParams) (*domain.ProductComponent, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, p)
	}
	items := make([]domain.ProductComponentItem, len(p.Items))
	for i, item := range p.Items {
		items[i] = domain.ProductComponentItem{
			ID: int64(i + 1), ProductComponentID: 1,
			ComponentProductItemID: item.ComponentProductItemID,
			Quantity:               item.Quantity,
			UnitID:                 item.UnitID,
		}
	}
	return &domain.ProductComponent{ID: 1, MerchantID: p.MerchantID, ProductItemID: p.ProductItemID, Items: items}, nil
}

func (m *mockComponentSvc) Delete(ctx context.Context, id int64, merchantID int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id, merchantID)
	}
	return nil
}

func withMerchant(c echo.Context) echo.Context {
	c.Set(httputil.ContextKeyMerchantID, int64(1))
	return c
}

func TestGetByProductItem(t *testing.T) {
	t.Parallel()

	t.Run("returns 200", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("productItemId")
		c.SetParamValues("1")
		c = withMerchant(c)

		h := NewHandler(&mockComponentSvc{})
		if err := h.GetByProductItem(c); err != nil {
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
		c.SetParamNames("productItemId")
		c.SetParamValues("999")
		c = withMerchant(c)

		h := NewHandler(&mockComponentSvc{
			getFn: func(_ context.Context, productItemID int64, merchantID int64) (*domain.ProductComponent, error) {
				return nil, domain.ErrProductComponentNotFound
			},
		})
		if err := h.GetByProductItem(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})
}

func TestCreateOrUpdate(t *testing.T) {
	t.Parallel()

	t.Run("create new component returns 201", func(t *testing.T) {
		e := echo.New()
		body := `{"items":[{"componentProductItemId":2,"quantity":2,"unitId":1}]}`
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("productItemId")
		c.SetParamValues("1")
		c = withMerchant(c)

		h := NewHandler(&mockComponentSvc{
			getFn: func(_ context.Context, productItemID int64, merchantID int64) (*domain.ProductComponent, error) {
				return nil, domain.ErrProductComponentNotFound
			},
		})
		if err := h.CreateOrUpdate(c); err != nil {
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

	t.Run("update existing component returns 200", func(t *testing.T) {
		e := echo.New()
		body := `{"items":[{"componentProductItemId":3,"quantity":5,"unitId":2}]}`
		req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("productItemId")
		c.SetParamValues("1")
		c = withMerchant(c)

		h := NewHandler(&mockComponentSvc{})
		if err := h.CreateOrUpdate(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp["data"] == nil {
			t.Error("expected data in response")
		}
	})
}
