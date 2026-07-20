package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

var errProviderTest = errors.New("provider error")

type mockStaffProvider struct {
	listByUserAndMerchant func(ctx context.Context, userID, merchantID int64) ([]StaffAssignment, error)
}

func (m *mockStaffProvider) ListByUserAndMerchant(ctx context.Context, userID, merchantID int64) ([]StaffAssignment, error) {
	return m.listByUserAndMerchant(ctx, userID, merchantID)
}

func TestBranchContext_FailClosedOnProviderError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/assignments", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set required context values that auth middleware normally provides
	c.Set("user_id", int64(1))
	c.Set("merchant_id", int64(42))

	provider := &mockStaffProvider{
		listByUserAndMerchant: func(_ context.Context, _, _ int64) ([]StaffAssignment, error) {
			return nil, errProviderTest
		},
	}

	middleware := BranchContext(provider)
	handler := middleware(func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("BranchContext returned an error to the next handler: %v", err)
	}

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rec.Code)
	}
}

func TestBranchContext_SuccessReturnsOK(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/assignments", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.Set("user_id", int64(1))
	c.Set("merchant_id", int64(42))

	provider := &mockStaffProvider{
		listByUserAndMerchant: func(_ context.Context, _, _ int64) ([]StaffAssignment, error) {
			return []StaffAssignment{
				{ID: 10, BranchID: 100, Role: "cashier", IsDefault: true},
			}, nil
		},
	}

	middleware := BranchContext(provider)
	handler := middleware(func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("BranchContext returned an error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestBranchContext_UnauthorizedWithoutUserID(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/me/assignments", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Do NOT set user_id — should fail with 401

	provider := &mockStaffProvider{
		listByUserAndMerchant: func(_ context.Context, _, _ int64) ([]StaffAssignment, error) {
			return nil, nil
		},
	}

	middleware := BranchContext(provider)
	handler := middleware(func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("BranchContext returned an error: %v", err)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}
