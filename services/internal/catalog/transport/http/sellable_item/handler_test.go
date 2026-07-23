package sellableitem

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
)

type mockItemSvc struct {
	createFn func(context.Context, usecase.CreateSellableItemParams) (*domain.SellableItem, error)
	listFn   func(context.Context, int64) ([]domain.SellableItem, error)
	updateFn func(context.Context, usecase.UpdateSellableItemParams) (*domain.SellableItem, error)
	deleteFn func(context.Context, int64) error
}

func (m *mockItemSvc) Create(ctx context.Context, p usecase.CreateSellableItemParams) (*domain.SellableItem, error) {
	if m.createFn != nil {
		return m.createFn(ctx, p)
	}
	return &domain.SellableItem{ID: 1, ProductID: p.ProductID, Name: p.Name, UnitID: p.UnitID, Status: domain.SellableItemStatusActive}, nil
}

func (m *mockItemSvc) ListByProduct(ctx context.Context, productID int64) ([]domain.SellableItem, error) {
	if m.listFn != nil {
		return m.listFn(ctx, productID)
	}
	return []domain.SellableItem{{ID: 1, ProductID: productID, Name: "Item", UnitID: 1}}, nil
}

func (m *mockItemSvc) Update(ctx context.Context, p usecase.UpdateSellableItemParams) (*domain.SellableItem, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, p)
	}
	return &domain.SellableItem{ID: p.ID, Name: "Updated", UnitID: 1}, nil
}

func (m *mockItemSvc) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockItemSvc) Restore(ctx context.Context, id int64) (*domain.SellableItem, error) {
	return &domain.SellableItem{ID: id, Name: "Restored", UnitID: 1, Price: 10000}, nil
}

func TestSellableItemCreate(t *testing.T) {
	t.Parallel()

	t.Run("valid request returns 201", func(t *testing.T) {
		e := echo.New()
		body := `{"name":"Marning Curah","unitId":1}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("productId")
		c.SetParamValues("1")

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

		h := NewHandler(&mockItemSvc{})
		if err := h.Create(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})
}

func TestSellableItemListByProduct(t *testing.T) {
	t.Parallel()

	t.Run("returns 200 with items", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("productId")
		c.SetParamValues("1")

		h := NewHandler(&mockItemSvc{})
		if err := h.ListByProduct(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("product not found returns 404", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("productId")
		c.SetParamValues("999")

		h := NewHandler(&mockItemSvc{
			listFn: func(_ context.Context, productID int64) ([]domain.SellableItem, error) {
				return nil, domain.ErrProductNotFound
			},
		})
		if err := h.ListByProduct(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})
}

func TestSellableItemUpdate(t *testing.T) {
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

		h := NewHandler(&mockItemSvc{})
		if err := h.Update(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestSellableItemDelete(t *testing.T) {
	t.Parallel()

	t.Run("success returns 200", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")

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

		h := NewHandler(&mockItemSvc{
			deleteFn: func(_ context.Context, id int64) error {
				return domain.ErrSellableItemNotFound
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
