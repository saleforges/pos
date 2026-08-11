package audit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
	"github.com/saleforge/pos/services/internal/iam/usecase"
	"github.com/saleforge/pos/services/pkg/pagination"
)

type mockAuthService struct {
	listFn func(ctx context.Context, userIDs []int64, p pagination.Params) ([]domain.LoginAudit, *pagination.Metadata, error)
}

func (m *mockAuthService) Register(ctx context.Context, params usecase.RegisterParams) (*usecase.AuthResult, error) {
	return nil, nil
}
func (m *mockAuthService) Login(ctx context.Context, params usecase.LoginParams) (*usecase.LoginResult, error) {
	return nil, nil
}
func (m *mockAuthService) RefreshToken(ctx context.Context, params usecase.RefreshTokenParams) (*usecase.LoginResult, error) {
	return nil, nil
}
func (m *mockAuthService) Logout(ctx context.Context, params usecase.LogoutParams) error { return nil }
func (m *mockAuthService) SwitchContext(ctx context.Context, sessionID string, userRoleID int64) (*usecase.AuthResult, error) {
	return nil, nil
}
func (m *mockAuthService) SetDefaultRole(ctx context.Context, userID, roleID int64) error { return nil }
func (m *mockAuthService) Introspect(ctx context.Context, tokenString string) (*usecase.IntrospectResult, error) {
	return nil, nil
}
func (m *mockAuthService) ValidateToken(ctx context.Context, tokenString string) (*port.TokenClaims, error) {
	return nil, nil
}
func (m *mockAuthService) HasPermission(claims *port.TokenClaims, required domain.Permission) bool {
	return true
}
func (m *mockAuthService) ListLoginAudits(ctx context.Context, userIDs []int64, p pagination.Params) ([]domain.LoginAudit, *pagination.Metadata, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userIDs, p)
	}
	return []domain.LoginAudit{}, &pagination.Metadata{}, nil
}

func TestHandler_ListLogins(t *testing.T) {
	t.Run("missing userIds returns 400", func(t *testing.T) {
		h := NewHandler(&mockAuthService{})
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/logins", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if err := h.ListLogins(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("invalid userIds returns 400", func(t *testing.T) {
		h := NewHandler(&mockAuthService{})
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/logins?userIds=abc", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if err := h.ListLogins(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("valid userIds passed through to usecase", func(t *testing.T) {
		var gotIDs []int64
		h := NewHandler(&mockAuthService{
			listFn: func(ctx context.Context, userIDs []int64, p pagination.Params) ([]domain.LoginAudit, *pagination.Metadata, error) {
				gotIDs = userIDs
				return []domain.LoginAudit{{ID: 1, UserID: 5}}, &pagination.Metadata{Count: 1}, nil
			},
		})
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/audit/logins?userIds=5,7", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		if err := h.ListLogins(c); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		if len(gotIDs) != 2 || gotIDs[0] != 5 || gotIDs[1] != 7 {
			t.Errorf("expected userIDs [5 7], got %v", gotIDs)
		}
	})
}
