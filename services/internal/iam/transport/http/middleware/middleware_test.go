package middleware

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
)

func TestBranchAccessMiddleware(t *testing.T) {
	t.Parallel()

	e := echo.New()
	e.GET("/branches/:branch_id/stock", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(claimsKey, &port.TokenClaims{UserID: 1})
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
		c.Set(claimsKey, &port.TokenClaims{UserID: 1})

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
				UserID:      1,
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

func TestCORSMiddleware(t *testing.T) {
	t.Parallel()

	e := echo.New()
	e.GET("/test", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, CORSMiddleware([]string{"http://localhost:3000"}))

	t.Run("allows configured origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		if rec.Header().Get("Access-Control-Allow-Origin") == "" {
			t.Error("expected CORS headers")
		}
	})

	t.Run("rejects disallowed origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "https://evil.com")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})

	t.Run("sets Vary: Origin header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Header().Get("Vary") != "Origin" {
			t.Errorf("expected Vary: Origin, got %q", rec.Header().Get("Vary"))
		}
	})
}

func TestLoginKey(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"username":"owner","password":"secret"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	key := LoginKey(c)
	if key != "owner@"+c.RealIP() {
		t.Fatalf("expected key owner@<ip>, got %q", key)
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		t.Fatalf("body not readable after LoginKey: %v", err)
	}
	if string(body) != `{"username":"owner","password":"secret"}` {
		t.Fatalf("body not restored, got %q", string(body))
	}
}

func TestLoginKeyFallsBackToIP(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`not-json`))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if key := LoginKey(c); key != c.RealIP() {
		t.Fatalf("expected IP fallback, got %q", key)
	}
}

func TestRateLimitMiddleware_PerAccountIsolation(t *testing.T) {
	t.Parallel()

	e := echo.New()
	e.POST("/auth/login", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, RateLimitMiddleware(1, time.Minute, LoginKey))

	login := func(username string) int {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(`{"username":"`+username+`","password":"x"}`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec.Code
	}

	if code := login("owner"); code != http.StatusOK {
		t.Fatalf("first owner login: expected 200, got %d", code)
	}
	if code := login("owner"); code != http.StatusTooManyRequests {
		t.Fatalf("second owner login: expected 429, got %d", code)
	}
	if code := login("cashier"); code != http.StatusOK {
		t.Fatalf("cashier login: expected 200 (different account not blocked), got %d", code)
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	t.Parallel()

	t.Run("allows request within limit", func(t *testing.T) {
		e := echo.New()
		e.GET("/test", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		}, RateLimitMiddleware(5, time.Minute, ClientIPKey))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("rejects request over limit", func(t *testing.T) {
		e := echo.New()
		e.GET("/test", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		}, RateLimitMiddleware(1, time.Minute, ClientIPKey))

		// First request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		// Second request — should be rate limited
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec2 := httptest.NewRecorder()
		e.ServeHTTP(rec2, req2)

		if rec2.Code != http.StatusTooManyRequests {
			t.Errorf("expected 429, got %d", rec2.Code)
		}
	})
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	t.Parallel()

	e := echo.New()
	e.GET("/me", func(c echo.Context) error {
		claims := c.Get(claimsKey).(*port.TokenClaims)
		if claims.UserID != 42 {
			t.Error("expected UserID 42 in context")
		}
		return c.NoContent(http.StatusOK)
	}, AuthMiddleware(func(_ context.Context, tokenString string) (*port.TokenClaims, error) {
		if tokenString != "valid-token" {
			return nil, domain.ErrInvalidToken
		}
		return &port.TokenClaims{UserID: 42, SessionID: "sess"}, nil
	}))

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestRBACMiddleware_AllowsWithPermission(t *testing.T) {
	t.Parallel()

	e := echo.New()
	e.GET("/admin", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(claimsKey, &port.TokenClaims{
				UserID:      1,
				Permissions: []domain.Permission{domain.UserDelete, domain.UserRead},
			})
			return next(c)
		}
	}, RBACMiddleware(domain.UserDelete, func(claims *port.TokenClaims, required domain.Permission) bool {
		for _, p := range claims.Permissions {
			if p == required {
				return true
			}
		}
		return false
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestBranchAccessMiddleware_MissingClaims(t *testing.T) {
	t.Parallel()

	handler := BranchAccessMiddleware("branch_id", func(_ *port.TokenClaims, branchID string) bool {
		return true
	})(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/branches/branch-1/stock", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/branches/:branch_id/stock")
	c.SetParamNames("branch_id")
	c.SetParamValues("branch-1")

	// Don't set claims — expect forbidden
	if err := handler(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}
