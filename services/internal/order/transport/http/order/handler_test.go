package order

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/order/domain"
	"github.com/saleforge/pos/services/internal/order/transport/http/middleware"
	"github.com/saleforge/pos/services/internal/order/usecase"
	"github.com/saleforge/pos/services/pkg/httputil"
)

type mockOrderSvc struct {
	createFn  func(context.Context, usecase.CreateOrderParams) (*domain.Order, error)
	getFn     func(context.Context, int64, int64) (*domain.Order, error)
	listFn    func(context.Context, int64, *int64, *domain.OrderStatus, *domain.PaymentStatus) ([]domain.Order, error)
	cancelFn  func(context.Context, int64, int64) (*domain.Order, error)
	paymentFn func(context.Context, usecase.AddPaymentParams) (*domain.Order, error)
}

func (m *mockOrderSvc) Create(ctx context.Context, p usecase.CreateOrderParams) (*domain.Order, error) {
	if m.createFn != nil {
		return m.createFn(ctx, p)
	}
	items := make([]domain.OrderItem, len(p.Items))
	for i, it := range p.Items {
		items[i] = domain.OrderItem{ID: int64(i + 1), OrderID: 1, ProductItemID: it.ProductItemID, ItemName: it.ItemName, UnitPrice: it.UnitPrice, Quantity: it.Quantity, LineTotal: it.UnitPrice * it.Quantity}
	}
	return &domain.Order{ID: 1, MerchantID: p.MerchantID, BranchID: p.BranchID, CreatedBy: p.CreatedBy, CustomerID: p.CustomerID, Status: domain.OrderStatusCompleted, PaymentStatus: domain.PaymentStatusUnpaid, Total: 30000, Items: items}, nil
}

func (m *mockOrderSvc) GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Order, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id, merchantID)
	}
	return &domain.Order{ID: id, MerchantID: merchantID, BranchID: 1, Status: domain.OrderStatusCompleted, PaymentStatus: domain.PaymentStatusPaid, Total: 30000}, nil
}

func (m *mockOrderSvc) List(ctx context.Context, merchantID int64, branchID *int64, status *domain.OrderStatus, paymentStatus *domain.PaymentStatus) ([]domain.Order, error) {
	if m.listFn != nil {
		return m.listFn(ctx, merchantID, branchID, status, paymentStatus)
	}
	return []domain.Order{{ID: 1, MerchantID: merchantID, BranchID: 1, Status: domain.OrderStatusCompleted, PaymentStatus: domain.PaymentStatusPaid, Total: 30000}}, nil
}

func (m *mockOrderSvc) Cancel(ctx context.Context, id int64, merchantID int64) (*domain.Order, error) {
	if m.cancelFn != nil {
		return m.cancelFn(ctx, id, merchantID)
	}
	return &domain.Order{ID: id, MerchantID: merchantID, BranchID: 1, Status: domain.OrderStatusCancelled, Total: 30000}, nil
}

func (m *mockOrderSvc) AddPayment(ctx context.Context, p usecase.AddPaymentParams) (*domain.Order, error) {
	if m.paymentFn != nil {
		return m.paymentFn(ctx, p)
	}
	return &domain.Order{ID: p.OrderID, MerchantID: p.MerchantID, BranchID: 1, Status: domain.OrderStatusCompleted, PaymentStatus: domain.PaymentStatusPaid, Total: 30000, PaidAmount: p.Amount}, nil
}

func withContext(c echo.Context) echo.Context {
	c.Set(httputil.ContextKeyMerchantID, int64(1))
	c.Set(middleware.ContextKeyUserID, int64(5))
	return c
}

func TestCreate(t *testing.T) {
	t.Parallel()

	t.Run("valid request returns 201 with createdBy from JWT", func(t *testing.T) {
		e := echo.New()
		body := `{"branchId":1,"items":[{"productItemId":35,"itemName":"Es Teh","unitPrice":15000,"quantity":2}]}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c = withContext(c)

		h := NewHandler(&mockOrderSvc{
			createFn: func(_ context.Context, p usecase.CreateOrderParams) (*domain.Order, error) {
				if p.CreatedBy != 5 {
					t.Errorf("expected createdBy 5 from JWT, got %d", p.CreatedBy)
				}
				if len(p.Items) != 1 {
					t.Errorf("expected 1 item, got %d", len(p.Items))
				}
				return &domain.Order{ID: 1, MerchantID: p.MerchantID, BranchID: p.BranchID, CreatedBy: p.CreatedBy, Status: domain.OrderStatusCompleted, PaymentStatus: domain.PaymentStatusUnpaid, Total: 30000}, nil
			},
		})
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
		c = withContext(c)

		h := NewHandler(&mockOrderSvc{
			createFn: func(_ context.Context, p usecase.CreateOrderParams) (*domain.Order, error) {
				if len(p.Items) == 0 {
					return nil, domain.ErrInvalidOrder
				}
				return &domain.Order{ID: 1}, nil
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
		c = withContext(c)

		h := NewHandler(&mockOrderSvc{})
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
		c = withContext(c)

		h := NewHandler(&mockOrderSvc{
			getFn: func(_ context.Context, id int64, merchantID int64) (*domain.Order, error) {
				return nil, domain.ErrOrderNotFound
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

	t.Run("returns 200 with filter", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/?branchId=1&paymentStatus=unpaid", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c = withContext(c)

		h := NewHandler(&mockOrderSvc{
			listFn: func(_ context.Context, merchantID int64, branchID *int64, status *domain.OrderStatus, paymentStatus *domain.PaymentStatus) ([]domain.Order, error) {
				if branchID == nil || *branchID != 1 {
					t.Error("expected branchID filter 1")
				}
				if paymentStatus == nil || *paymentStatus != domain.PaymentStatusUnpaid {
					t.Error("expected paymentStatus filter unpaid")
				}
				return []domain.Order{}, nil
			},
		})
		if err := h.List(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}

func TestCancel(t *testing.T) {
	t.Parallel()

	t.Run("valid cancel returns 200", func(t *testing.T) {
		e := echo.New()
		body := `{"status":"cancelled"}`
		req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")
		c = withContext(c)

		h := NewHandler(&mockOrderSvc{})
		if err := h.Cancel(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("non-cancelled status returns 400", func(t *testing.T) {
		e := echo.New()
		body := `{"status":"completed"}`
		req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")
		c = withContext(c)

		h := NewHandler(&mockOrderSvc{})
		if err := h.Cancel(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})
}

func TestAddPayment(t *testing.T) {
	t.Parallel()

	t.Run("valid payment returns 200", func(t *testing.T) {
		e := echo.New()
		body := `{"amount":20000,"method":"cash"}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues("1")
		c = withContext(c)

		h := NewHandler(&mockOrderSvc{
			paymentFn: func(_ context.Context, p usecase.AddPaymentParams) (*domain.Order, error) {
				if p.CreatedBy != 5 {
					t.Errorf("expected createdBy 5 from JWT, got %d", p.CreatedBy)
				}
				if p.Method != domain.PaymentMethodCash {
					t.Errorf("expected method cash, got %s", p.Method)
				}
				return &domain.Order{ID: p.OrderID, MerchantID: p.MerchantID, Status: domain.OrderStatusCompleted, PaymentStatus: domain.PaymentStatusPartial, Total: 30000, PaidAmount: p.Amount}, nil
			},
		})
		if err := h.AddPayment(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})
}
