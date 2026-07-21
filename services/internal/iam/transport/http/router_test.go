package httptransport

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
	"github.com/saleforge/pos/services/internal/iam/transport/http/auth"
	"github.com/saleforge/pos/services/internal/iam/transport/http/permission"
	"github.com/saleforge/pos/services/internal/iam/transport/http/role"
	"github.com/saleforge/pos/services/internal/iam/transport/http/user"
	"github.com/saleforge/pos/services/internal/iam/usecase"
)

// mockUserRepo implements repository.UserRepository with context.Context
type mockUserRepo struct {
	existing map[string]bool
}

func (m *mockUserRepo) Create(_ context.Context, user *domain.User) error {
	user.ID = 1
	return nil
}
func (m *mockUserRepo) GetByID(_ context.Context, id int64) (*domain.User, error) {
	return &domain.User{ID: id, Username: "test", Email: "test@test.com", Status: domain.UserStatusActive}, nil
}
func (m *mockUserRepo) GetByUsername(_ context.Context, username string) (*domain.User, error) {
	if username == "existinguser" {
		return &domain.User{ID: 1, Username: "existinguser"}, nil
	}
	return nil, domain.ErrUserNotFound
}
func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	if email == "existing@test.com" {
		return &domain.User{ID: 1, Email: "existing@test.com"}, nil
	}
	return nil, domain.ErrUserNotFound
}
func (m *mockUserRepo) List(_ context.Context, offset, limit int) ([]domain.User, error) {
	return []domain.User{}, nil
}
func (m *mockUserRepo) Update(_ context.Context, user *domain.User) error          { return nil }
func (m *mockUserRepo) Delete(_ context.Context, id int64) error                    { return nil }
func (m *mockUserRepo) AddRole(_ context.Context, userID int64, roleName string) error  { return nil }
func (m *mockUserRepo) RemoveRole(_ context.Context, userID int64, roleName string) error { return nil }

type mockRoleRepo struct{}

func (m *mockRoleRepo) Create(_ context.Context, role *domain.Role) error {
	role.ID = 1
	return nil
}
func (m *mockRoleRepo) GetByID(_ context.Context, id int64) (*domain.Role, error) {
	return &domain.Role{ID: id, Name: "test_role"}, nil
}
func (m *mockRoleRepo) GetByName(_ context.Context, name string) (*domain.Role, error) {
	return &domain.Role{ID: 1, Name: name, Permissions: []domain.Permission{domain.UserRead}}, nil
}
func (m *mockRoleRepo) List(_ context.Context, _ *int64) ([]domain.Role, error)    { return nil, nil }
func (m *mockRoleRepo) Update(_ context.Context, role *domain.Role) error          { return nil }
func (m *mockRoleRepo) Delete(_ context.Context, id int64) error                    { return nil }
func (m *mockRoleRepo) AddPermission(_ context.Context, roleID int64, permission domain.Permission) error {
	return nil
}
func (m *mockRoleRepo) RemovePermission(_ context.Context, roleID int64, permission domain.Permission) error {
	return nil
}
func (m *mockRoleRepo) GetPermissions(_ context.Context, roleID int64) ([]domain.Permission, error) {
	return nil, nil
}

type mockPermissionRepo struct{}

func (m *mockPermissionRepo) Create(_ context.Context, p domain.Permission) error   { return nil }
func (m *mockPermissionRepo) GetAll(_ context.Context) ([]domain.Permission, error) { return nil, nil }
func (m *mockPermissionRepo) Delete(_ context.Context, p domain.Permission) error   { return nil }

type mockLoginAuditRepo struct{}

func (m *mockLoginAuditRepo) Create(_ context.Context, audit *domain.LoginAudit) error { return nil }
func (m *mockLoginAuditRepo) List(_ context.Context, offset, limit int) ([]domain.LoginAudit, error) {
	return nil, nil
}

type mockStaffRepo struct{}

func (m *mockStaffRepo) ListByUserID(_ context.Context, userID int64) ([]domain.UserRoleAssignment, error) {
	return nil, nil
}
func (m *mockStaffRepo) SetDefaultRole(_ context.Context, userID, roleID int64) error { return nil }
func (m *mockStaffRepo) Create(_ context.Context, userID int64, merchantID int64, merchantName, role string) error {
	return nil
}

type mockSessionStore struct{}

func (m *mockSessionStore) Create(_ context.Context, session *domain.Session) error { return nil }
func (m *mockSessionStore) Get(_ context.Context, id string) (*domain.Session, error) {
	return &domain.Session{ID: id}, nil
}
func (m *mockSessionStore) Update(_ context.Context, session *domain.Session) error { return nil }
func (m *mockSessionStore) Delete(_ context.Context, id string) error                { return nil }

type mockEventPublisher struct{}

func (m *mockEventPublisher) Publish(_ context.Context, _ string, _ interface{}) error { return nil }

type mockTokenHasher struct{}

func (m *mockTokenHasher) HashToken(token string) string {
	if token == "" { return "" }
	return "hashed:" + token
}

type mockPasswordHasher struct{}

func (m *mockPasswordHasher) Hash(password string) (string, error) { return "hashed:" + password, nil }
func (m *mockPasswordHasher) Compare(hashedPassword, password string) error {
	if hashedPassword != "hashed:"+password {
		return domain.ErrInvalidCredentials
	}
	return nil
}

type mockTokenSigner struct{}

func (m *mockTokenSigner) SignAccessToken(_ port.TokenClaims) (string, error) {
	return "access-token", nil
}
func (m *mockTokenSigner) SignRefreshToken(_ int64, _ string) (string, error) {
	return "refresh-token", nil
}
func (m *mockTokenSigner) VerifyAccessToken(_ string) (*port.TokenClaims, error) {
	return &port.TokenClaims{UserID: 1, SessionID: "test-session"}, nil
}
func (m *mockTokenSigner) VerifyRefreshToken(_ string) (int64, string, error) {
	return 1, "test-session", nil
}

func TestHandler_Routes_Respond(t *testing.T) {
	t.Parallel()

	authService := usecase.NewAuthUsecase(
		&mockUserRepo{},
		&mockRoleRepo{},
		&mockPermissionRepo{},
		&mockLoginAuditRepo{},
		&mockStaffRepo{},
		&mockSessionStore{},
		&mockEventPublisher{},
		&mockPasswordHasher{},
		&mockTokenSigner{},
		&mockTokenHasher{},
		nil,
		nil,
	)

	userService := usecase.NewUserUsecase(
		&mockUserRepo{},
		&mockStaffRepo{},
		&mockEventPublisher{},
		nil,
	)

	roleService := usecase.NewRoleUsecase(
		&mockRoleRepo{},
		&mockUserRepo{},
	)

	authHandler := auth.NewHandler(authService, userService, true)
	userHandler := user.NewHandler(authService, userService)
	roleHandler := role.NewHandler(roleService)
	permHandler := permission.NewHandler(usecase.NewPermissionUsecase(&mockPermissionRepo{}))

	e := echo.New()

	t.Run("GET /api/v1/auth/me returns unauthorized without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/auth/me")

		authHandler.Me(c)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("POST /api/v1/auth/register with valid body succeeds", func(t *testing.T) {
		body := `{"username":"newuser","email":"new@test.com","password":"Secure1pass"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/auth/register")

		authHandler.Register(c)

		if rec.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("POST /api/v1/auth/register with empty body returns 400", func(t *testing.T) {
		body := `{}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/auth/register")

		authHandler.Register(c)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("POST /api/v1/auth/login returns 401 for invalid credentials", func(t *testing.T) {
		body := `{"username":"nonexistent","password":"wrong"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/auth/login")

		authHandler.Login(c)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("GET /api/v1/users returns user list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/users")

		userHandler.ListUsers(c)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("GET /api/v1/roles returns role list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/roles", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/roles")

		roleHandler.ListRoles(c)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("GET /api/v1/permissions returns permission list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/permissions", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/permissions")

		permHandler.ListPermissions(c)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("POST /api/v1/roles creates a new role", func(t *testing.T) {
		body := `{"name":"test_role","description":"test desc","permissions":["user.read"]}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/roles", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/v1/roles")

		roleHandler.CreateRole(c)

		if rec.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}

		var wrapped struct {
			Data domain.Role `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &wrapped); err != nil {
			t.Fatalf("failed to decode response: %v\nbody: %s", err, rec.Body.String())
		}
		if wrapped.Data.Name != "test_role" {
			t.Errorf("expected name 'test_role', got %q", wrapped.Data.Name)
		}
	})
}
