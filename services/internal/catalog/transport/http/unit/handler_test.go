package unit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type mockUnitSvc struct {
	listFn func(context.Context) ([]domain.Unit, error)
}

func (m *mockUnitSvc) List(ctx context.Context) ([]domain.Unit, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return []domain.Unit{
		{ID: 1, Code: "PCS", Name: "Piece"},
		{ID: 2, Code: "KG", Name: "Kilogram"},
	}, nil
}

func TestUnitList(t *testing.T) {
	t.Parallel()

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := NewHandler(&mockUnitSvc{})
	if err := h.List(c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
