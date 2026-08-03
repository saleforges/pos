package customer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/order/domain"
	"github.com/saleforge/pos/services/internal/order/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
)

type mockCustomerSvc struct {
	createFn func(context.Context, usecase.CreateCustomerParams) (*domain.Customer, error)
	getFn    func(context.Context, int64, int64) (*domain.Customer, error)
	listFn   func(context.Context, int64, string) ([]domain.Customer, error)
	updateFn func(context.Context, usecase.UpdateCustomerParams) (*domain.Customer, error)
	deleteFn func(context.Context, int64, int64) error
}

func (m *mockCustomerSvc) Create(ctx context.Context, p usecase.CreateCustomerParams) (*domain.Customer, error) {
	if m.createFn != nil {
		return m.createFn(ctx, p)
	}
	return &domain.Customer{ID: 1, MerchantID: p.MerchantID, Name: p.Name, Phone: p.Phone}, nil
}

func (m *mockCustomerSvc) GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Customer, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id, merchantID)
	}
	return &domain.Customer{ID: id, MerchantID: merchantID, Name: "Pak Budi"}, nil
}

func (m *mockCustomerSvc) List(ctx context.Context, merchantID int64, search string) ([]domain.Customer, error) {
	if m.listFn != nil {
		return m.listFn(ctx, merchantID, search)
	}
	return []domain.Customer{{ID: 1, MerchantID: merchantID, Name: "Pak Budi"}}, nil
}

func (m *mockCustomerSvc) Update(ctx context.Context, p usecase.UpdateCustomerParams) (*domain.Customer, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, p)
	}
	return &domain.Customer{ID: p.ID, MerchantID: p.MerchantID, Name: "Updated"}, nil
}

func (m *mockCustomerSvc) Delete(ctx context.Context, id int64, merchantID int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id, merchantID)
	}
	return nil
}

func (m *mockCustomerSvc) Sync(ctx context.Context, merchantID int64, lastSync *time.Time) (*usecase.CustomerSyncResult, error) {
	return &usecase.CustomerSyncResult{Customers: []domain.Customer{}, SyncToken: "2026-08-03T00:00:00Z"}, nil
}

func withMerchant(c echo.Context) echo.Context {
	c.Set(httputil.ContextKeyMerchantID, int64(1))
	return c
}

func TestCreate(t *testing.T) {
	t.Parallel()

	t.Run("valid request returns 201", func(t *testing.T) {
		e := echo.New()
		body := `{"name":"Pak Budi","phone":"0812345"}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c = withMerchant(c)

		h := NewHandler(&mockCustomerSvc{})
		if err := h.Create(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", rec.Code)
		}
	})

	t.Run("empty name returns 400", func(t *testing.T) {
		e := echo.New()
		body := `{"name":""}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c = withMerchant(c)

		h := NewHandler(&mockCustomerSvc{
			createFn: func(_ context.Context, p usecase.CreateCustomerParams) (*domain.Customer, error) {
				if p.Name == "" {
					return nil, domain.ErrInvalidCustomer
				}
				return &domain.Customer{ID: 1}, nil
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

	t.Run("not found returns 404", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("999")
		c = withMerchant(c)

		h := NewHandler(&mockCustomerSvc{
			getFn: func(_ context.Context, id int64, merchantID int64) (*domain.Customer, error) {
				return nil, domain.ErrCustomerNotFound
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

	t.Run("returns 200", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/?q=budi", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c = withMerchant(c)

		h := NewHandler(&mockCustomerSvc{})
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
		body := `{"phone":"081999"}`
		req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")
		c = withMerchant(c)

		h := NewHandler(&mockCustomerSvc{})
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

	t.Run("returns 200", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")
		c = withMerchant(c)

		h := NewHandler(&mockCustomerSvc{})
		if err := h.Delete(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}
