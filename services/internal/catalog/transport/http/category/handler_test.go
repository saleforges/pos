package category

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

type mockCategorySvc struct {
	createFn func(context.Context, usecase.CreateCategoryParams) (*domain.Category, error)
	getFn    func(context.Context, int64) (*domain.Category, error)
	listFn   func(context.Context, int64) ([]domain.Category, error)
	updateFn func(context.Context, usecase.UpdateCategoryParams) (*domain.Category, error)
	deleteFn func(context.Context, int64) error
}

func (m *mockCategorySvc) Create(ctx context.Context, p usecase.CreateCategoryParams) (*domain.Category, error) {
	if m.createFn != nil {
		return m.createFn(ctx, p)
	}
	return &domain.Category{ID: 1, MerchantID: p.MerchantID, Name: p.Name}, nil
}

func (m *mockCategorySvc) GetByID(ctx context.Context, id int64) (*domain.Category, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id)
	}
	return &domain.Category{ID: id, MerchantID: 1, Name: "Test"}, nil
}

func (m *mockCategorySvc) ListByMerchant(ctx context.Context, merchantID int64) ([]domain.Category, error) {
	if m.listFn != nil {
		return m.listFn(ctx, merchantID)
	}
	return []domain.Category{{ID: 1, MerchantID: merchantID, Name: "Snack"}}, nil
}

func (m *mockCategorySvc) Update(ctx context.Context, p usecase.UpdateCategoryParams) (*domain.Category, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, p)
	}
	return &domain.Category{ID: p.ID, Name: "Updated"}, nil
}

func (m *mockCategorySvc) Delete(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func (m *mockCategorySvc) Restore(ctx context.Context, id int64) (*domain.Category, error) {
	return &domain.Category{ID: id, Name: "Restored"}, nil
}

func withMerchant(c echo.Context) echo.Context {
	c.Set(httputil.ContextKeyMerchantID, int64(1))
	return c
}

func TestCategoryCreate(t *testing.T) {
	t.Parallel()

	t.Run("valid request returns 201", func(t *testing.T) {
		e := echo.New()
		body := `{"name":"Snack"}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := withMerchant(e.NewContext(req, rec))

		h := NewHandler(&mockCategorySvc{})
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
		body := `{"name":""}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := withMerchant(e.NewContext(req, rec))

		h := NewHandler(&mockCategorySvc{})
		if err := h.Create(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})
}

func TestCategoryList(t *testing.T) {
	t.Parallel()

	t.Run("returns 200 with categories", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := withMerchant(e.NewContext(req, rec))

		h := NewHandler(&mockCategorySvc{})
		if err := h.List(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestCategoryUpdate(t *testing.T) {
	t.Parallel()

	t.Run("returns 200", func(t *testing.T) {
		e := echo.New()
		body := `{"name":"Renamed"}`
		req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")

		h := NewHandler(&mockCategorySvc{})
		if err := h.Update(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("not found returns 404", func(t *testing.T) {
		e := echo.New()
		body := `{"name":"Renamed"}`
		req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("999")

		h := NewHandler(&mockCategorySvc{
			updateFn: func(_ context.Context, p usecase.UpdateCategoryParams) (*domain.Category, error) {
				return nil, domain.ErrCategoryNotFound
			},
		})
		if err := h.Update(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", rec.Code)
		}
	})
}
