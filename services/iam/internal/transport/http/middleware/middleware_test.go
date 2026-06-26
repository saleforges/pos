package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/iam/internal/domain"
	"github.com/saleforge/pos/services/iam/internal/port"
)

func TestBranchAccessMiddleware(t *testing.T) {
	t.Parallel()

	e := echo.New()
	e.GET("/branches/:branch_id/stock", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(claimsKey, &port.TokenClaims{UserID: "u1"})
			return next(c)
		}
	}, BranchAccessMiddleware("branch_id", func(_ *port.TokenClaims, branchID string) bool {
		return branchID == "branch-1"
	}))

	t.Run("allows matching branch", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/branches/branch-1/stock", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})

	t.Run("rejects missing branch", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/branches/stock", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/branches/:branch_id/stock")
		c.SetParamNames("branch_id")
		c.SetParamValues("")
		c.Set(claimsKey, &port.TokenClaims{UserID: "u1"})

		handler := BranchAccessMiddleware("branch_id", func(_ *port.TokenClaims, branchID string) bool {
			return branchID == "branch-1"
		})(func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		if err := handler(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("rejects non matching branch", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/branches/branch-2/stock", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
		}
	})
}

func TestAuthMiddlewareRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	e := echo.New()
	e.GET("/me", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, AuthMiddleware(func(context.Context, string) (*port.TokenClaims, error) {
		return nil, errors.New("should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestRBACMiddlewareRejectsMissingPermission(t *testing.T) {
	t.Parallel()

	e := echo.New()
	e.GET("/admin", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(claimsKey, &port.TokenClaims{
				UserID:      "u1",
				Permissions: []domain.Permission{domain.UserRead},
			})
			return next(c)
		}
	}, RBACMiddleware(domain.UserDelete, func(claims *port.TokenClaims, required domain.Permission) bool {
		for _, permission := range claims.Permissions {
			if permission == required {
				return true
			}
		}
		return false
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}
