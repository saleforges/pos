package httputil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestRequirePermission(t *testing.T) {
	e := echo.New()
	called := false
	handler := RequirePermission(PermCatalogCreate)(func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusOK)
	})

	run := func(perms []string) (*httptest.ResponseRecorder, error) {
		called = false
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		if perms != nil {
			c.Set(ContextKeyPermissions, perms)
		}
		err := handler(c)
		return rec, err
	}

	t.Run("allows when permission present", func(t *testing.T) {
		rec, err := run([]string{"catalog.read", PermCatalogCreate})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !called {
			t.Fatal("expected next handler to be called")
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("blocks when permission missing", func(t *testing.T) {
		rec, err := run([]string{"catalog.read"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if called {
			t.Fatal("expected next handler NOT to be called")
		}
		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("blocks when no permissions set on context", func(t *testing.T) {
		rec, err := run(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if called {
			t.Fatal("expected next handler NOT to be called")
		}
		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", rec.Code)
		}
	})
}
