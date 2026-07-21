package user

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
	"github.com/saleforge/pos/services/internal/iam/usecase"
)

type mockAuthSvc struct {
	registerFn   func(ctx context.Context, params usecase.RegisterParams) (*usecase.AuthResult, error)
}

func (m *mockAuthSvc) Register(ctx context.Context, params usecase.RegisterParams) (*usecase.AuthResult, error) {
	if m.registerFn != nil { return m.registerFn(ctx, params) }
	return &usecase.AuthResult{TokenPair: port.TokenPair{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 3600}}, nil
}
func (m *mockAuthSvc) Login(ctx context.Context, params usecase.LoginParams) (*usecase.LoginResult, error) {
	return &usecase.LoginResult{TokenPair: port.TokenPair{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 3600}}, nil
}
func (m *mockAuthSvc) RefreshToken(ctx context.Context, params usecase.RefreshTokenParams) (*usecase.LoginResult, error) {
	return &usecase.LoginResult{TokenPair: port.TokenPair{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 3600}}, nil
}
func (m *mockAuthSvc) Logout(ctx context.Context, params usecase.LogoutParams) error { return nil }
func (m *mockAuthSvc) SwitchContext(ctx context.Context, sessionID string, userRoleID int64) (*usecase.AuthResult, error) {
	return &usecase.AuthResult{TokenPair: port.TokenPair{AccessToken: "at", ExpiresIn: 3600}}, nil
}
func (m *mockAuthSvc) SetDefaultRole(ctx context.Context, userID, roleID int64) error { return nil }
func (m *mockAuthSvc) Introspect(ctx context.Context, tokenString string) (*usecase.IntrospectResult, error) {
	return &usecase.IntrospectResult{Active: true, UserID: 1}, nil
}
func (m *mockAuthSvc) ValidateToken(ctx context.Context, tokenString string) (*port.TokenClaims, error) {
	return &port.TokenClaims{UserID: 1, SessionID: "sess"}, nil
}
func (m *mockAuthSvc) HasPermission(claims *port.TokenClaims, required domain.Permission) bool { return true }

type mockUserSvc struct {
	listUsersFn  func(ctx context.Context, offset, limit int) ([]domain.User, error)
	getUserFn    func(ctx context.Context, id int64) (*domain.User, error)
	updateUserFn func(ctx context.Context, params usecase.UpdateUserParams) (*domain.User, error)
	deleteUserFn func(ctx context.Context, id int64) error
}

func (m *mockUserSvc) ListUsers(ctx context.Context, offset, limit int) ([]domain.User, error) {
	if m.listUsersFn != nil { return m.listUsersFn(ctx, offset, limit) }
	return []domain.User{{ID: 1, Username: "u1"}, {ID: 2, Username: "u2"}}, nil
}
func (m *mockUserSvc) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	if m.getUserFn != nil { return m.getUserFn(ctx, id) }
	return &domain.User{ID: id, Username: "test", Email: "t@t.com", Status: domain.UserStatusActive}, nil
}
func (m *mockUserSvc) UpdateUser(ctx context.Context, params usecase.UpdateUserParams) (*domain.User, error) {
	if m.updateUserFn != nil { return m.updateUserFn(ctx, params) }
	return &domain.User{ID: params.ID, Username: *params.Username}, nil
}
func (m *mockUserSvc) DeleteUser(ctx context.Context, id int64) error {
	if m.deleteUserFn != nil { return m.deleteUserFn(ctx, id) }
	return nil
}
func (m *mockUserSvc) ListStaff(ctx context.Context, userID int64) ([]domain.UserRoleAssignment, error) { return nil, nil }

func newTestHandler(us usecase.UserUsecase) *Handler {
	return NewHandler(&mockAuthSvc{}, us)
}

func TestListUsers(t *testing.T) {
	t.Parallel()

	h := newTestHandler(&mockUserSvc{})
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/users")

	h.ListUsers(c)
	if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }

	var wrapped struct {
		Data []domain.User `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &wrapped); err != nil {
		t.Fatalf("bad response: %v", err)
	}
	if len(wrapped.Data) != 2 { t.Errorf("expected 2 users, got %d", len(wrapped.Data)) }
}

func TestGetUser(t *testing.T) {
	t.Parallel()

	t.Run("existing user returns 200", func(t *testing.T) {
		h := newTestHandler(&mockUserSvc{})
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/users/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		h.GetUser(c)
		if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
	})

	t.Run("non-existent user returns 404", func(t *testing.T) {
		us := &mockUserSvc{getUserFn: func(_ context.Context, _ int64) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		}}
		h := newTestHandler(us)
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/users/999", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/users/:id")
		c.SetParamNames("id")
		c.SetParamValues("999")

		h.GetUser(c)
		if rec.Code != http.StatusNotFound { t.Errorf("expected 404, got %d", rec.Code) }
	})

	t.Run("invalid id returns 400", func(t *testing.T) {
		h := newTestHandler(&mockUserSvc{})
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/users/abc", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/users/:id")
		c.SetParamNames("id")
		c.SetParamValues("abc")

		h.GetUser(c)
		if rec.Code != http.StatusBadRequest { t.Errorf("expected 400, got %d", rec.Code) }
	})
}

func TestCreateUser(t *testing.T) {
	t.Parallel()

	t.Run("valid request returns 201", func(t *testing.T) {
		h := newTestHandler(&mockUserSvc{})
		e := echo.New()
		body := `{"username":"newu","email":"new@t.com","password":"Secure1pass","roles":["viewer"]}`
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/users")

		h.CreateUser(c)
		if rec.Code != http.StatusCreated { t.Errorf("expected 201, got %d: %s", rec.Code, rec.Body.String()) }
	})

	t.Run("password policy error maps to 400", func(t *testing.T) {
		h := NewHandler(&mockAuthSvc{registerFn: func(_ context.Context, _ usecase.RegisterParams) (*usecase.AuthResult, error) {
			return nil, domain.ErrPasswordPolicy
		}}, &mockUserSvc{})
		e := echo.New()
		body := `{"username":"newu","email":"new@t.com","password":"weak"}`
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/users")

		h.CreateUser(c)
		if rec.Code != http.StatusBadRequest { t.Errorf("expected 400, got %d", rec.Code) }
	})

	t.Run("duplicate user maps to 409", func(t *testing.T) {
		h := NewHandler(&mockAuthSvc{registerFn: func(_ context.Context, _ usecase.RegisterParams) (*usecase.AuthResult, error) {
			return nil, domain.ErrUserAlreadyExists
		}}, &mockUserSvc{})
		e := echo.New()
		body := `{"username":"dup","email":"dup@t.com","password":"Secure1pass"}`
		req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/users")

		h.CreateUser(c)
		if rec.Code != http.StatusConflict { t.Errorf("expected 409, got %d", rec.Code) }
	})
}

func TestUpdateUser(t *testing.T) {
	t.Parallel()

	t.Run("valid update returns 200", func(t *testing.T) {
		h := newTestHandler(&mockUserSvc{})
		e := echo.New()
		body := `{"username":"updated"}`
		req := httptest.NewRequest(http.MethodPatch, "/users/1", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/users/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		h.UpdateUser(c)
		if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
	})

	t.Run("user not found maps to 404", func(t *testing.T) {
		us := &mockUserSvc{updateUserFn: func(_ context.Context, _ usecase.UpdateUserParams) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		}}
		h := newTestHandler(us)
		e := echo.New()
		body := `{"username":"updated"}`
		req := httptest.NewRequest(http.MethodPatch, "/users/999", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/users/:id")
		c.SetParamNames("id")
		c.SetParamValues("999")

		h.UpdateUser(c)
		if rec.Code != http.StatusNotFound { t.Errorf("expected 404, got %d", rec.Code) }
	})
}

func TestDeleteUser(t *testing.T) {
	t.Parallel()

	t.Run("existing user returns 200", func(t *testing.T) {
		h := newTestHandler(&mockUserSvc{})
		e := echo.New()
		req := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/users/:id")
		c.SetParamNames("id")
		c.SetParamValues("1")

		h.DeleteUser(c)
		if rec.Code != http.StatusOK { t.Errorf("expected 200, got %d", rec.Code) }
	})
}
