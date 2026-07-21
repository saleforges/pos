package permission

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/domain"
)

type mockPermissionService struct {
	listFn   func(ctx context.Context) ([]domain.Permission, error)
	createFn func(ctx context.Context, p domain.Permission) error
	deleteFn func(ctx context.Context, p domain.Permission) error
}

func (m *mockPermissionService) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	if m.listFn != nil { return m.listFn(ctx) }
	return []domain.Permission{domain.UserRead, domain.UserCreate}, nil
}
func (m *mockPermissionService) CreatePermission(ctx context.Context, p domain.Permission) error {
	if m.createFn != nil { return m.createFn(ctx, p) }
	return nil
}
func (m *mockPermissionService) DeletePermission(ctx context.Context, p domain.Permission) error {
	if m.deleteFn != nil { return m.deleteFn(ctx, p) }
	return nil
}

func TestListPermissions(t *testing.T) {
	t.Parallel()

	h := NewHandler(&mockPermissionService{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/permissions", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/permissions")

	h.ListPermissions(c)
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }

	var wrapped struct { Data []domain.Permission `json:"data"` }
	if err := json.Unmarshal(rec.Body.Bytes(), &wrapped); err != nil {
		t.Fatalf("bad response: %v", err)
	}
	if len(wrapped.Data) != 2 { t.Errorf("expected 2 permissions, got %d", len(wrapped.Data)) }
}

func TestCreatePermission(t *testing.T) {
	t.Parallel()

	t.Run("valid request returns 201", func(t *testing.T) {
		h := NewHandler(&mockPermissionService{})
		e := echo.New()
		body := `{"permission":"reports.read"}`
		req := httptest.NewRequest(http.MethodPost, "/permissions", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/permissions")

		h.CreatePermission(c)
		if rec.Code != http.StatusCreated { t.Errorf("expected 201, got %d", rec.Code) }
	})

	t.Run("invalid body returns 400", func(t *testing.T) {
		h := NewHandler(&mockPermissionService{})
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/permissions", strings.NewReader(`not-json`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/permissions")

		h.CreatePermission(c)
		if rec.Code != http.StatusBadRequest { t.Errorf("expected 400, got %d", rec.Code) }
	})
}

func TestDeletePermission(t *testing.T) {
	t.Parallel()

	h := NewHandler(&mockPermissionService{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/permissions/reports.read", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/permissions/:name")
	c.SetParamNames("name")
	c.SetParamValues("reports.read")

	h.DeletePermission(c)
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}
