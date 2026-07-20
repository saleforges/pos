package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/internal/merchant/usecase"
)

type mockBranchSvc struct {
	createBranch func(context.Context, usecase.CreateBranchInput) (*domain.Branch, error)
	getBranch    func(context.Context, int64) (*domain.Branch, error)
	listBranches func(context.Context, int64) ([]domain.Branch, error)
	updateBranch func(context.Context, usecase.UpdateBranchInput) (*domain.Branch, error)
	deleteBranch func(context.Context, int64) error
}

func (m *mockBranchSvc) CreateBranch(ctx context.Context, i usecase.CreateBranchInput) (*domain.Branch, error) {
	return m.createBranch(ctx, i)
}
func (m *mockBranchSvc) GetBranch(ctx context.Context, id int64) (*domain.Branch, error) {
	return m.getBranch(ctx, id)
}
func (m *mockBranchSvc) ListBranches(ctx context.Context, merchantID int64) ([]domain.Branch, error) {
	return m.listBranches(ctx, merchantID)
}
func (m *mockBranchSvc) UpdateBranch(ctx context.Context, i usecase.UpdateBranchInput) (*domain.Branch, error) {
	return m.updateBranch(ctx, i)
}
func (m *mockBranchSvc) DeleteBranch(ctx context.Context, id int64) error {
	return m.deleteBranch(ctx, id)
}

func TestBranchHandler_CreateBranch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		mock       *mockBranchSvc
		wantStatus int
		wantBody   string
	}{
		{
			name: "success",
			body: `{"name":"Downtown","code":"DTC-01","address":"123 Main St"}`,
			mock: &mockBranchSvc{
				createBranch: func(_ context.Context, i usecase.CreateBranchInput) (*domain.Branch, error) {
					return &domain.Branch{
						ID: 1, MerchantID: i.MerchantID, Name: i.Name,
						Code: i.Code, Status: domain.BranchStatusActive,
					}, nil
				},
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"name":"Downtown"`,
		},
		{
			name:       "invalid json",
			body:       `{bad}`,
			mock:       &mockBranchSvc{},
			wantStatus: http.StatusBadRequest,
			wantBody:   `"message":"invalid request body"`,
		},
		{
			name:       "missing fields",
			body:       `{"name":"test"}`,
			mock:       &mockBranchSvc{},
			wantStatus: http.StatusBadRequest,
			wantBody:   `"message":"missing required fields"`,
		},
		{
			name: "branch exists",
			body: `{"name":"Downtown","code":"DTC-01","address":"123 Main St"}`,
			mock: &mockBranchSvc{
				createBranch: func(_ context.Context, _ usecase.CreateBranchInput) (*domain.Branch, error) {
					return nil, domain.ErrBranchExists
				},
			},
			wantStatus: http.StatusConflict,
		},
		{
			name: "invalid branch from usecase",
			body: `{"name":"Downtown","code":"DTC-01","address":"123 Main St"}`,
			mock: &mockBranchSvc{
				createBranch: func(_ context.Context, _ usecase.CreateBranchInput) (*domain.Branch, error) {
					return nil, domain.ErrInvalidBranch
				},
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/branches", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.Set("merchant_id", int64(1))

			h := NewBranchHandler(tt.mock)
			err := h.CreateBranch(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
			if tt.wantBody != "" && !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Errorf("expected body to contain %q, got %s", tt.wantBody, rec.Body.String())
			}
		})
	}
}

func TestBranchHandler_GetBranch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		mock       *mockBranchSvc
		wantStatus int
	}{
		{
			name: "success",
			id:   "1",
			mock: &mockBranchSvc{
				getBranch: func(_ context.Context, id int64) (*domain.Branch, error) {
					return &domain.Branch{ID: id, Name: "Downtown", Status: domain.BranchStatusActive}, nil
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			id:   "999",
			mock: &mockBranchSvc{
				getBranch: func(_ context.Context, _ int64) (*domain.Branch, error) {
					return nil, domain.ErrBranchNotFound
				},
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/branches/"+tt.id, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			h := NewBranchHandler(tt.mock)
			err := h.GetBranch(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestBranchHandler_DeleteBranch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		mock       *mockBranchSvc
		wantStatus int
	}{
		{
			name: "success",
			id:   "1",
			mock: &mockBranchSvc{deleteBranch: func(_ context.Context, _ int64) error { return nil }},
			wantStatus: http.StatusOK,
		},
		{
			name: "internal error",
			id:   "1",
			mock: &mockBranchSvc{deleteBranch: func(_ context.Context, _ int64) error { return domain.ErrInternal }},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodDelete, "/api/v1/branches/"+tt.id, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			h := NewBranchHandler(tt.mock)
			err := h.DeleteBranch(c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}
