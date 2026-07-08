package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/internal/merchant/usecase"
)

type mockMerchantSvc struct {
	createMerchantFn func(context.Context, usecase.CreateMerchantInput) (*domain.Merchant, error)
	getMerchantFn    func(context.Context, string) (*domain.Merchant, error)
	listMerchantsFn  func(context.Context, int, int) ([]domain.Merchant, error)
	updateMerchantFn func(context.Context, usecase.UpdateMerchantInput) (*domain.Merchant, error)
	deleteMerchantFn func(context.Context, string) error
}

func (m *mockMerchantSvc) CreateMerchant(ctx context.Context, i usecase.CreateMerchantInput) (*domain.Merchant, error) {
	return m.createMerchantFn(ctx, i)
}
func (m *mockMerchantSvc) GetMerchant(ctx context.Context, id string) (*domain.Merchant, error) {
	return m.getMerchantFn(ctx, id)
}
func (m *mockMerchantSvc) ListMerchants(ctx context.Context, offset, limit int) ([]domain.Merchant, error) {
	return m.listMerchantsFn(ctx, offset, limit)
}
func (m *mockMerchantSvc) UpdateMerchant(ctx context.Context, i usecase.UpdateMerchantInput) (*domain.Merchant, error) {
	return m.updateMerchantFn(ctx, i)
}
func (m *mockMerchantSvc) DeleteMerchant(ctx context.Context, id string) error {
	return m.deleteMerchantFn(ctx, id)
}

func now() time.Time {
	return time.Date(2026, 7, 2, 0, 0, 0, 0, time.UTC)
}

func TestMerchantHandler_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		mock       *mockMerchantSvc
		wantStatus int
		wantBody   string
	}{
		{
			name: "success",
			body: `{"name":"Toko Maju","email":"toko@maju.com","legal_name":"PT Maju","address":"Jl. Merdeka","phone":"021-1234","tax_id":"01.234.567.8-999.000"}`,
			mock: &mockMerchantSvc{
				createMerchantFn: func(_ context.Context, i usecase.CreateMerchantInput) (*domain.Merchant, error) {
					return &domain.Merchant{
						ID: "m1", Name: i.Name, LegalName: i.LegalName,
						Address: i.Address, Phone: i.Phone, Email: i.Email, TaxID: i.TaxID,
						Status: domain.MerchantStatusActive, CreatedAt: now(), UpdatedAt: now(),
					}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"Toko Maju"`,
		},
		{
			name:       "invalid json",
			body:       `{invalid}`,
			mock:       &mockMerchantSvc{},
			wantStatus: http.StatusBadRequest,
			wantBody:   `"message":"invalid request body"`,
		},
		{
			name:       "missing required fields",
			body:       `{"name":""}`,
			mock:       &mockMerchantSvc{},
			wantStatus: http.StatusBadRequest,
			wantBody:   `"message":"missing required fields"`,
		},
		{
			name: "internal error",
			body: `{"name":"Toko Maju","email":"toko@maju.com"}`,
			mock: &mockMerchantSvc{
				createMerchantFn: func(_ context.Context, _ usecase.CreateMerchantInput) (*domain.Merchant, error) {
					return nil, domain.ErrInternal
				},
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `"message":"MCH500: internal error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/merchants", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			h := NewMerchantHandler(tt.mock)
			err := h.Create(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("expected body to contain %q, got %s", tt.wantBody, rec.Body.String())
			}
		})
	}
}

func TestMerchantHandler_Get(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		mock       *mockMerchantSvc
		wantStatus int
		wantBody   string
	}{
		{
			name: "success",
			id:   "m1",
			mock: &mockMerchantSvc{
				getMerchantFn: func(_ context.Context, id string) (*domain.Merchant, error) {
					return &domain.Merchant{ID: id, Name: "Toko Maju", Status: domain.MerchantStatusActive}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"Toko Maju"`,
		},
		{
			name: "not found",
			id:   "nonexistent",
			mock: &mockMerchantSvc{
				getMerchantFn: func(_ context.Context, _ string) (*domain.Merchant, error) {
					return nil, domain.ErrMerchantNotFound
				},
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `"message":"MCH001: merchant not found"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/merchants/"+tt.id, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			h := NewMerchantHandler(tt.mock)
			err := h.Get(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("expected body to contain %q, got %s", tt.wantBody, rec.Body.String())
			}
		})
	}
}

func TestMerchantHandler_List(t *testing.T) {
	t.Parallel()

	mock := &mockMerchantSvc{
		listMerchantsFn: func(_ context.Context, offset, limit int) ([]domain.Merchant, error) {
			return []domain.Merchant{
				{ID: "m1", Name: "Toko Maju"},
				{ID: "m2", Name: "Toko Baru"},
			}, nil
		},
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/merchants?offset=0&limit=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewMerchantHandler(mock)
	err := h.List(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var wrap struct {
		Message string            `json:"message"`
		Data    []domain.Merchant `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &wrap); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(wrap.Data) != 2 {
		t.Errorf("expected 2 merchants, got %d", len(wrap.Data))
	}
}

func TestMerchantHandler_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		body       string
		mock       *mockMerchantSvc
		wantStatus int
		wantBody   string
	}{
		{
			name: "success",
			id:   "m1",
			body: `{"name":"Updated Store"}`,
			mock: &mockMerchantSvc{
				updateMerchantFn: func(_ context.Context, i usecase.UpdateMerchantInput) (*domain.Merchant, error) {
					return &domain.Merchant{ID: i.ID, Name: *i.Name, Status: domain.MerchantStatusActive}, nil
				},
			},
			wantStatus: http.StatusOK,
			wantBody:   `"name":"Updated Store"`,
		},
		{
			name:       "invalid json",
			id:         "m1",
			body:       `{bad}`,
			mock:       &mockMerchantSvc{},
			wantStatus: http.StatusBadRequest,
			wantBody:   `"message":"invalid request body"`,
		},
		{
			name: "not found",
			id:   "nonexistent",
			body: `{"name":"Updated Store"}`,
			mock: &mockMerchantSvc{
				updateMerchantFn: func(_ context.Context, _ usecase.UpdateMerchantInput) (*domain.Merchant, error) {
					return nil, domain.ErrMerchantNotFound
				},
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `"message":"MCH001: merchant not found"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/merchants/"+tt.id, strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			h := NewMerchantHandler(tt.mock)
			err := h.Update(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
			if !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("expected body to contain %q, got %s", tt.wantBody, rec.Body.String())
			}
		})
	}
}

func TestMerchantHandler_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		mock       *mockMerchantSvc
		wantStatus int
	}{
		{
			name: "success",
			id:   "m1",
			mock: &mockMerchantSvc{
				deleteMerchantFn: func(_ context.Context, id string) error {
					return nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "internal error",
			id:   "m1",
			mock: &mockMerchantSvc{
				deleteMerchantFn: func(_ context.Context, _ string) error {
					return domain.ErrInternal
				},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/merchants/"+tt.id, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			h := NewMerchantHandler(tt.mock)
			err := h.Delete(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}
