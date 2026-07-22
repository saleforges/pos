package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
	"github.com/saleforge/pos/services/internal/iam/transport/http/common"
	"github.com/saleforge/pos/services/internal/iam/usecase"
	"github.com/saleforge/pos/services/pkg/pagination"
)

// mockAuthService implements usecase.AuthService
type mockAuthService struct {
	registerFn    func(ctx context.Context, params usecase.RegisterParams) (*usecase.AuthResult, error)
	loginFn       func(ctx context.Context, params usecase.LoginParams) (*usecase.LoginResult, error)
	refreshFn     func(ctx context.Context, params usecase.RefreshTokenParams) (*usecase.LoginResult, error)
	logoutFn      func(ctx context.Context, params usecase.LogoutParams) error
	switchCtxFn   func(ctx context.Context, sessionID string, userRoleID int64) (*usecase.AuthResult, error)
	setDefaultFn  func(ctx context.Context, userID, roleID int64) error
	introspectFn  func(ctx context.Context, tokenString string) (*usecase.IntrospectResult, error)
	validateFn    func(ctx context.Context, tokenString string) (*port.TokenClaims, error)
	hasPermFn     func(claims *port.TokenClaims, required domain.Permission) bool
}

func (m *mockAuthService) Register(ctx context.Context, params usecase.RegisterParams) (*usecase.AuthResult, error) {
	if m.registerFn != nil {
		return m.registerFn(ctx, params)
	}
	return &usecase.AuthResult{TokenPair: port.TokenPair{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 3600}}, nil
}
func (m *mockAuthService) Login(ctx context.Context, params usecase.LoginParams) (*usecase.LoginResult, error) {
	if m.loginFn != nil { return m.loginFn(ctx, params) }
	return &usecase.LoginResult{TokenPair: port.TokenPair{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 3600}}, nil
}
func (m *mockAuthService) RefreshToken(ctx context.Context, params usecase.RefreshTokenParams) (*usecase.LoginResult, error) {
	if m.refreshFn != nil { return m.refreshFn(ctx, params) }
	return &usecase.LoginResult{TokenPair: port.TokenPair{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 3600}}, nil
}
func (m *mockAuthService) Logout(ctx context.Context, params usecase.LogoutParams) error {
	if m.logoutFn != nil { return m.logoutFn(ctx, params) }
	return nil
}
func (m *mockAuthService) SwitchContext(ctx context.Context, sessionID string, userRoleID int64) (*usecase.AuthResult, error) {
	if m.switchCtxFn != nil { return m.switchCtxFn(ctx, sessionID, userRoleID) }
	return &usecase.AuthResult{TokenPair: port.TokenPair{AccessToken: "at", ExpiresIn: 3600}}, nil
}
func (m *mockAuthService) SetDefaultRole(ctx context.Context, userID, roleID int64) error {
	if m.setDefaultFn != nil { return m.setDefaultFn(ctx, userID, roleID) }
	return nil
}
func (m *mockAuthService) Introspect(ctx context.Context, tokenString string) (*usecase.IntrospectResult, error) {
	if m.introspectFn != nil { return m.introspectFn(ctx, tokenString) }
	return &usecase.IntrospectResult{Active: true, UserID: 1}, nil
}
func (m *mockAuthService) ValidateToken(ctx context.Context, tokenString string) (*port.TokenClaims, error) {
	if m.validateFn != nil { return m.validateFn(ctx, tokenString) }
	return &port.TokenClaims{UserID: 1, SessionID: "sess"}, nil
}
func (m *mockAuthService) HasPermission(claims *port.TokenClaims, required domain.Permission) bool {
	if m.hasPermFn != nil { return m.hasPermFn(claims, required) }
	return true
}

type mockUserService struct {
	getUserFn func(ctx context.Context, id int64) (*domain.User, error)
}

func (m *mockUserService) Register(ctx context.Context, params usecase.RegisterParams) (*usecase.AuthResult, error) {
	return &usecase.AuthResult{TokenPair: port.TokenPair{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 3600}}, nil
}
func (m *mockUserService) ListUsers(ctx context.Context, p pagination.Params) ([]domain.User, *pagination.Metadata, error) { return nil, nil, nil }
func (m *mockUserService) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	if m.getUserFn != nil { return m.getUserFn(ctx, id) }
	return &domain.User{ID: id, Username: "test", Email: "t@t.com", Status: domain.UserStatusActive}, nil
}
func (m *mockUserService) UpdateUser(ctx context.Context, params usecase.UpdateUserParams) (*domain.User, error) {
	return &domain.User{ID: params.ID}, nil
}
func (m *mockUserService) DeleteUser(ctx context.Context, id int64) error { return nil }
func (m *mockUserService) AssignRole(ctx context.Context, userID int64, roleName string) error { return nil }
func (m *mockUserService) RemoveRole(ctx context.Context, userID int64, roleName string) error { return nil }
func (m *mockUserService) ListStaff(ctx context.Context, userID int64) ([]domain.UserRoleAssignment, error) { return nil, nil }

func setupTest(t *testing.T, as *mockAuthService, us *mockUserService) (*Handler, echo.Context, *httptest.ResponseRecorder) {
	t.Helper()
	if as == nil { as = &mockAuthService{} }
	if us == nil { us = &mockUserService{} }
	return NewHandler(as, us, true), echo.New().NewContext(nil, nil), httptest.NewRecorder()
}

func request(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return req
}

func withClaims(c echo.Context) echo.Context {
	c.Set(common.ClaimsKey, &port.TokenClaims{UserID: 1, SessionID: "sess-1"})
	return c
}

func TestRegister(t *testing.T) {
	t.Parallel()

	t.Run("valid request returns 201", func(t *testing.T) {
		h, _, rec := setupTest(t, nil, nil)
		c := echo.New().NewContext(request(http.MethodPost, "/auth/register", `{"username":"u","email":"u@t.com","password":"Secure1pass"}`), rec)
		c.SetPath("/auth/register")

		err := h.Register(c)
		if err != nil { t.Fatalf("unexpected error: %v", err) }
		if rec.Code != http.StatusCreated { t.Errorf("expected 201, got %d", rec.Code) }

		var wrapped struct { Data authResponse `json:"data"` }
		if err := json.Unmarshal(rec.Body.Bytes(), &wrapped); err != nil {
			t.Fatalf("bad response: %v", err)
		}
		if wrapped.Data.AccessToken == "" { t.Error("expected access token") }
	})

	t.Run("empty body returns 400", func(t *testing.T) {
		h, _, rec := setupTest(t, nil, nil)
		c := echo.New().NewContext(request(http.MethodPost, "/auth/register", `{}`), rec)
		c.SetPath("/auth/register")

		h.Register(c)
		if rec.Code != http.StatusBadRequest { t.Errorf("expected 400, got %d", rec.Code) }
	})

	t.Run("missing fields returns 400", func(t *testing.T) {
		h, _, rec := setupTest(t, nil, nil)
		c := echo.New().NewContext(request(http.MethodPost, "/auth/register", `{"username":"u"}`), rec)
		c.SetPath("/auth/register")

		h.Register(c)
		if rec.Code != http.StatusBadRequest { t.Errorf("expected 400, got %d", rec.Code) }
	})

	t.Run("password policy error maps to 400", func(t *testing.T) {
		as := &mockAuthService{registerFn: func(_ context.Context, _ usecase.RegisterParams) (*usecase.AuthResult, error) {
			return nil, domain.ErrPasswordPolicy
		}}
		h, _, rec := setupTest(t, as, nil)
		c := echo.New().NewContext(request(http.MethodPost, "/auth/register", `{"username":"u","email":"u@t.com","password":"weak"}`), rec)
		c.SetPath("/auth/register")

		h.Register(c)
		if rec.Code != http.StatusBadRequest { t.Errorf("expected 400, got %d", rec.Code) }
	})

	t.Run("duplicate user maps to 409", func(t *testing.T) {
		as := &mockAuthService{registerFn: func(_ context.Context, _ usecase.RegisterParams) (*usecase.AuthResult, error) {
			return nil, domain.ErrUserAlreadyExists
		}}
		h, _, rec := setupTest(t, as, nil)
		c := echo.New().NewContext(request(http.MethodPost, "/auth/register", `{"username":"u","email":"u@t.com","password":"Secure1pass"}`), rec)
		c.SetPath("/auth/register")

		h.Register(c)
		if rec.Code != http.StatusConflict { t.Errorf("expected 409, got %d", rec.Code) }
	})

	t.Run("internal error maps to 500", func(t *testing.T) {
		as := &mockAuthService{registerFn: func(_ context.Context, _ usecase.RegisterParams) (*usecase.AuthResult, error) {
			return nil, domain.ErrInternal
		}}
		h, _, rec := setupTest(t, as, nil)
		c := echo.New().NewContext(request(http.MethodPost, "/auth/register", `{"username":"u","email":"u@t.com","password":"Secure1pass"}`), rec)
		c.SetPath("/auth/register")

		h.Register(c)
		if rec.Code != http.StatusInternalServerError { t.Errorf("expected 500, got %d", rec.Code) }
	})
}

func TestLogin(t *testing.T) {
	t.Parallel()

	t.Run("valid login returns 200", func(t *testing.T) {
		h, _, rec := setupTest(t, nil, nil)
		c := echo.New().NewContext(request(http.MethodPost, "/auth/login", `{"username":"u","password":"pass"}`), rec)
		c.SetPath("/auth/login")

		h.Login(c)
		if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
	})

	t.Run("missing fields returns 400", func(t *testing.T) {
		h, _, rec := setupTest(t, nil, nil)
		c := echo.New().NewContext(request(http.MethodPost, "/auth/login", `{}`), rec)
		c.SetPath("/auth/login")

		h.Login(c)
		if rec.Code != http.StatusBadRequest { t.Errorf("expected 400, got %d", rec.Code) }
	})

	t.Run("invalid credentials maps to 401", func(t *testing.T) {
		as := &mockAuthService{loginFn: func(_ context.Context, _ usecase.LoginParams) (*usecase.LoginResult, error) {
			return nil, domain.ErrInvalidCredentials
		}}
		h, _, rec := setupTest(t, as, nil)
		c := echo.New().NewContext(request(http.MethodPost, "/auth/login", `{"username":"u","password":"wrong"}`), rec)
		c.SetPath("/auth/login")

		h.Login(c)
		if rec.Code != http.StatusUnauthorized { t.Errorf("expected 401, got %d", rec.Code) }
	})

	t.Run("disabled user maps to 401", func(t *testing.T) {
		as := &mockAuthService{loginFn: func(_ context.Context, _ usecase.LoginParams) (*usecase.LoginResult, error) {
			return nil, domain.ErrUserDisabled
		}}
		h, _, rec := setupTest(t, as, nil)
		c := echo.New().NewContext(request(http.MethodPost, "/auth/login", `{"username":"u","password":"pass"}`), rec)
		c.SetPath("/auth/login")

		h.Login(c)
		if rec.Code != http.StatusUnauthorized { t.Errorf("expected 401, got %d", rec.Code) }
	})
}

func TestRefresh(t *testing.T) {
	t.Parallel()

	t.Run("valid body returns 200", func(t *testing.T) {
		h, _, rec := setupTest(t, nil, nil)
		c := echo.New().NewContext(request(http.MethodPost, "/auth/refresh", `{"refresh_token":"rt"}`), rec)
		c.SetPath("/auth/refresh")

		h.Refresh(c)
		if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
	})

	t.Run("missing refresh_token returns 400", func(t *testing.T) {
		h, _, rec := setupTest(t, nil, nil)
		c := echo.New().NewContext(request(http.MethodPost, "/auth/refresh", `{}`), rec)
		c.SetPath("/auth/refresh")

		h.Refresh(c)
		if rec.Code != http.StatusBadRequest { t.Errorf("expected 400, got %d", rec.Code) }
	})

	t.Run("invalid token maps to 401", func(t *testing.T) {
		as := &mockAuthService{refreshFn: func(_ context.Context, _ usecase.RefreshTokenParams) (*usecase.LoginResult, error) {
			return nil, domain.ErrInvalidRefreshToken
		}}
		h, _, rec := setupTest(t, as, nil)
		c := echo.New().NewContext(request(http.MethodPost, "/auth/refresh", `{"refresh_token":"bad"}`), rec)
		c.SetPath("/auth/refresh")

		h.Refresh(c)
		if rec.Code != http.StatusUnauthorized { t.Errorf("expected 401, got %d", rec.Code) }
	})

	t.Run("disabled user maps to 403", func(t *testing.T) {
		as := &mockAuthService{refreshFn: func(_ context.Context, _ usecase.RefreshTokenParams) (*usecase.LoginResult, error) {
			return nil, domain.ErrUserDisabled
		}}
		h, _, rec := setupTest(t, as, nil)
		c := echo.New().NewContext(request(http.MethodPost, "/auth/refresh", `{"refresh_token":"rt"}`), rec)
		c.SetPath("/auth/refresh")

		h.Refresh(c)
		if rec.Code != http.StatusForbidden { t.Errorf("expected 403, got %d", rec.Code) }
	})
}

func TestLogout(t *testing.T) {
	t.Parallel()

	t.Run("successful logout returns 200", func(t *testing.T) {
		h, _, rec := setupTest(t, nil, nil)
		c := echo.New().NewContext(request(http.MethodPost, "/auth/logout", ""), rec)
		c.SetPath("/auth/logout")
		withClaims(c)

		h.Logout(c)
		if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
	})

	t.Run("service error returns 500", func(t *testing.T) {
		as := &mockAuthService{logoutFn: func(_ context.Context, _ usecase.LogoutParams) error {
			return domain.ErrInternal
		}}
		h, _, rec := setupTest(t, as, nil)
		c := echo.New().NewContext(request(http.MethodPost, "/auth/logout", ""), rec)
		c.SetPath("/auth/logout")
		withClaims(c)

		h.Logout(c)
		if rec.Code != http.StatusInternalServerError { t.Errorf("expected 500, got %d", rec.Code) }
	})
}

func TestSwitchContext(t *testing.T) {
	t.Parallel()

	t.Run("valid request returns 200", func(t *testing.T) {
		h, _, rec := setupTest(t, nil, nil)
		c := echo.New().NewContext(request(http.MethodPost, "/auth/switch-context", `{"userRoleId":1}`), rec)
		c.SetPath("/auth/switch-context")
		withClaims(c)

		h.SwitchContext(c)
		if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
	})

	t.Run("session not found maps to 403", func(t *testing.T) {
		as := &mockAuthService{switchCtxFn: func(_ context.Context, _ string, _ int64) (*usecase.AuthResult, error) {
			return nil, domain.ErrSessionNotFound
		}}
		h, _, rec := setupTest(t, as, nil)
		c := echo.New().NewContext(request(http.MethodPost, "/auth/switch-context", `{"userRoleId":1}`), rec)
		c.SetPath("/auth/switch-context")
		withClaims(c)

		h.SwitchContext(c)
		if rec.Code != http.StatusForbidden { t.Errorf("expected 403, got %d", rec.Code) }
	})
}

func TestSetDefaultRole(t *testing.T) {
	t.Parallel()

	h, _, rec := setupTest(t, nil, nil)
	c := echo.New().NewContext(request(http.MethodPut, "/auth/me/default-role", `{"userRoleId":1}`), rec)
	c.SetPath("/auth/me/default-role")
	withClaims(c)

	h.SetDefaultRole(c)
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
}

func TestIntrospect(t *testing.T) {
	t.Parallel()

	t.Run("valid token returns 200 with active", func(t *testing.T) {
		h, _, rec := setupTest(t, nil, nil)
		c := echo.New().NewContext(request(http.MethodPost, "/auth/introspect", `{"token":"valid"}`), rec)
		c.SetPath("/auth/introspect")

		h.Introspect(c)
		if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }

		var wrapped struct { Data *usecase.IntrospectResult `json:"data"` }
		if err := json.Unmarshal(rec.Body.Bytes(), &wrapped); err != nil {
			t.Fatalf("bad response: %v", err)
		}
		if wrapped.Data == nil || !wrapped.Data.Active { t.Error("expected active result") }
	})

	t.Run("service error returns 500", func(t *testing.T) {
		as := &mockAuthService{introspectFn: func(_ context.Context, _ string) (*usecase.IntrospectResult, error) {
			return nil, domain.ErrInternal
		}}
		h, _, rec := setupTest(t, as, nil)
		c := echo.New().NewContext(request(http.MethodPost, "/auth/introspect", `{"token":"bad"}`), rec)
		c.SetPath("/auth/introspect")

		h.Introspect(c)
		if rec.Code != http.StatusInternalServerError { t.Errorf("expected 500, got %d", rec.Code) }
	})
}

func TestMe(t *testing.T) {
	t.Parallel()

	t.Run("authenticated returns 200", func(t *testing.T) {
		h, _, rec := setupTest(t, nil, nil)
		c := echo.New().NewContext(request(http.MethodGet, "/auth/me", ""), rec)
		c.SetPath("/auth/me")
		withClaims(c)

		h.Me(c)
		if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
	})

	t.Run("no claims returns 401", func(t *testing.T) {
		h, _, rec := setupTest(t, nil, nil)
		c := echo.New().NewContext(request(http.MethodGet, "/auth/me", ""), rec)
		c.SetPath("/auth/me")

		h.Me(c)
		if rec.Code != http.StatusUnauthorized { t.Errorf("expected 401, got %d", rec.Code) }
	})

	t.Run("user not found returns 404", func(t *testing.T) {
		us := &mockUserService{getUserFn: func(_ context.Context, _ int64) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		}}
		h, _, rec := setupTest(t, nil, us)
		c := echo.New().NewContext(request(http.MethodGet, "/auth/me", ""), rec)
		c.SetPath("/auth/me")
		withClaims(c)

		h.Me(c)
		if rec.Code != http.StatusNotFound { t.Errorf("expected 404, got %d", rec.Code) }
	})
}
