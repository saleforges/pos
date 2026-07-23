package product

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
	"github.com/saleforge/pos/services/pkg/pagination"
)

type mockCategoryRepo struct{}

func (m *mockCategoryRepo) Create(_ context.Context, _ *domain.Category) error { return nil }
func (m *mockCategoryRepo) GetByID(_ context.Context, id int64) (*domain.Category, error) { return nil, domain.ErrCategoryNotFound }
func (m *mockCategoryRepo) ListByMerchant(_ context.Context, _ int64) ([]domain.Category, error) { return nil, nil }
func (m *mockCategoryRepo) Update(_ context.Context, _ *domain.Category) error { return nil }
func (m *mockCategoryRepo) Delete(_ context.Context, _ int64) error { return nil }
func (m *mockCategoryRepo) Restore(_ context.Context, _ int64) (*domain.Category, error) { return nil, nil }

type mockSellableItemRepo struct{}

func (m *mockSellableItemRepo) Create(_ context.Context, _ *domain.SellableItem) error { return nil }
func (m *mockSellableItemRepo) GetByID(_ context.Context, _ int64) (*domain.SellableItem, error) { return nil, domain.ErrSellableItemNotFound }
func (m *mockSellableItemRepo) ListByProduct(_ context.Context, _ int64) ([]domain.SellableItem, error) { return nil, nil }
func (m *mockSellableItemRepo) Update(_ context.Context, _ *domain.SellableItem) error { return nil }
func (m *mockSellableItemRepo) Delete(_ context.Context, _ int64) error { return nil }
func (m *mockSellableItemRepo) Restore(_ context.Context, _ int64) (*domain.SellableItem, error) { return nil, nil }

type mockUnitRepo struct{}

func (m *mockUnitRepo) GetAll(_ context.Context) ([]domain.Unit, error) { return nil, nil }
func (m *mockUnitRepo) GetByID(_ context.Context, _ int64) (*domain.Unit, error) { return nil, domain.ErrUnitNotFound }
func (m *mockUnitRepo) GetByCode(_ context.Context, _ string) (*domain.Unit, error) { return nil, domain.ErrUnitNotFound }

type mockProductSvc struct {
	createFn func(context.Context, usecase.CreateProductParams) (*domain.Product, error)
	getFn    func(context.Context, int64) (*domain.Product, error)
	listFn   func(context.Context, int64, string, pagination.Params) ([]domain.Product, *pagination.Metadata, error)
	updateFn func(context.Context, usecase.UpdateProductParams) (*domain.Product, error)
	deleteFn func(context.Context, int64) error
}

func (m *mockProductSvc) Create(ctx context.Context, p usecase.CreateProductParams) (*domain.Product, error) {
	if m.createFn != nil {
		return m.createFn(ctx, p)
	}
	return &domain.Product{ID: 1, MerchantID: p.MerchantID, CategoryID: p.CategoryID, Name: p.Name, Status: domain.ProductStatusActive}, nil
}

func (m *mockProductSvc) GetByID(ctx context.Context, id int64) (*domain.Product, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id)
	}
	return &domain.Product{ID: id, MerchantID: 1, CategoryID: 1, Name: "Test", Status: domain.ProductStatusActive}, nil
}

func (m *mockProductSvc) List(ctx context.Context, merchantID int64, search string, p pagination.Params) ([]domain.Product, *pagination.Metadata, error) {
	if m.listFn != nil {
		return m.listFn(ctx, merchantID, search, p)
	}
	return []domain.Product{{ID: 1, MerchantID: merchantID, Name: "Test"}}, &pagination.Metadata{Total: 1, Offset: 0, Limit: 20, Count: 1}, nil
}

func (m *mockProductSvc) Update(ctx context.Context, p usecase.UpdateProductParams) (*domain.Product, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, p)
	}
	return &domain.Product{ID: p.ID, Name: "Updated"}, nil
}

func (m *mockProductSvc) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func withMerchant(c echo.Context) echo.Context {
	c.Set(httputil.ContextKeyMerchantID, int64(1))
	return c
}

func TestCreate(t *testing.T) {
	t.Parallel()

	t.Run("valid request returns 201", func(t *testing.T) {
		e := echo.New()
		body := `{"name":"Marning","categoryId":1}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := withMerchant(e.NewContext(req, rec))

		h := NewHandler(&mockProductSvc{}, &mockCategoryRepo{}, &mockSellableItemRepo{}, &mockUnitRepo{})
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
		body := `{"name":"","categoryId":1}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := withMerchant(e.NewContext(req, rec))

		h := NewHandler(&mockProductSvc{}, &mockCategoryRepo{}, &mockSellableItemRepo{}, &mockUnitRepo{})
		if err := h.Create(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("missing fields returns 400", func(t *testing.T) {
		e := echo.New()
		body := `{}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := withMerchant(e.NewContext(req, rec))

		h := NewHandler(&mockProductSvc{}, &mockCategoryRepo{}, &mockSellableItemRepo{}, &mockUnitRepo{})
		if err := h.Create(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})
}

func TestGet(t *testing.T) {
	t.Parallel()

	t.Run("found returns 200", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")

		h := NewHandler(&mockProductSvc{}, &mockCategoryRepo{}, &mockSellableItemRepo{}, &mockUnitRepo{})
		if err := h.Get(c); err != nil {
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

		mock := &mockProductSvc{
			getFn: func(_ context.Context, id int64) (*domain.Product, error) {
				return nil, domain.ErrProductNotFound
			},
		}
		h := NewHandler(mock, &mockCategoryRepo{}, &mockSellableItemRepo{}, &mockUnitRepo{})
		if err := h.Get(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})
}

func TestList(t *testing.T) {
	t.Parallel()

	t.Run("returns 200 with pagination", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/?offset=0&limit=10", nil)
		rec := httptest.NewRecorder()
		c := withMerchant(e.NewContext(req, rec))

		h := NewHandler(&mockProductSvc{}, &mockCategoryRepo{}, &mockSellableItemRepo{}, &mockUnitRepo{})
		if err := h.List(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		if resp["pagination"] == nil {
			t.Error("expected pagination in response")
		}
	})
}

func TestUpdate(t *testing.T) {
	t.Parallel()

	t.Run("valid request returns 200", func(t *testing.T) {
		e := echo.New()
		body := `{"name":"Updated Name"}`
		req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")

		h := NewHandler(&mockProductSvc{}, &mockCategoryRepo{}, &mockSellableItemRepo{}, &mockUnitRepo{})
		if err := h.Update(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("not found returns 404", func(t *testing.T) {
		e := echo.New()
		body := `{"name":"New"}`
		req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("999")

		mock := &mockProductSvc{
			updateFn: func(_ context.Context, p usecase.UpdateProductParams) (*domain.Product, error) {
				return nil, domain.ErrProductNotFound
			},
		}
		h := NewHandler(mock, &mockCategoryRepo{}, &mockSellableItemRepo{}, &mockUnitRepo{})
		if err := h.Update(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
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

		h := NewHandler(&mockProductSvc{}, &mockCategoryRepo{}, &mockSellableItemRepo{}, &mockUnitRepo{})
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

		mock := &mockProductSvc{
			deleteFn: func(_ context.Context, id int64) error {
				return domain.ErrProductNotFound
			},
		}
		h := NewHandler(mock, &mockCategoryRepo{}, &mockSellableItemRepo{}, &mockUnitRepo{})
		if err := h.Delete(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})
}
