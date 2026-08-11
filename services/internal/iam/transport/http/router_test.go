package httptransport

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
	"github.com/saleforge/pos/services/internal/iam/transport/http/audit"
	"github.com/saleforge/pos/services/internal/iam/transport/http/auth"
	"github.com/saleforge/pos/services/internal/iam/transport/http/permission"
	"github.com/saleforge/pos/services/internal/iam/transport/http/role"
	"github.com/saleforge/pos/services/internal/iam/transport/http/user"
	"github.com/saleforge/pos/services/internal/iam/usecase"
	"github.com/saleforge/pos/services/pkg/pagination"
)

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
func (m *mockUserRepo) List(_ context.Context, offset, limit int) ([]domain.User, int64, error) {
	return []domain.User{}, 0, nil
}
func (m *mockUserRepo) Update(_ context.Context, user *domain.User) error                 { return nil }
func (m *mockUserRepo) Delete(_ context.Context, id int64) error                          { return nil }
func (m *mockUserRepo) AddRole(_ context.Context, userID int64, roleName string) error    { return nil }
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
func (m *mockRoleRepo) List(_ context.Context, _ *int64) ([]domain.Role, error) { return nil, nil }
func (m *mockRoleRepo) Update(_ context.Context, role *domain.Role) error       { return nil }
func (m *mockRoleRepo) Delete(_ context.Context, id int64) error                { return nil }
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
func (m *mockLoginAuditRepo) List(_ context.Context, userIDs []int64, offset, limit int) ([]domain.LoginAudit, int64, error) {
	return nil, 0, nil
}

type mockStaffRepo struct{}

func (m *mockStaffRepo) ListByUserID(_ context.Context, userID int64) ([]domain.UserRoleAssignment, error) {
	return nil, nil
}
func (m *mockStaffRepo) SetDefaultRole(_ context.Context, userID, roleID int64) error { return nil }
func (m *mockStaffRepo) Create(_ context.Context, userID int64, merchantID int64, merchantName, role string) error {
	return nil
}
func (m *mockStaffRepo) CreateWithTx(_ context.Context, tx port.Tx, userID int64, merchantID int64, merchantName, role string) error {
	return nil
}

func TestHTTPRouterSmoke(t *testing.T) {
	e := NewRouter(
		auth.NewHandler(&mockAuthUsecase{}, &mockUserUsecase{}, false),
		user.NewHandler(&mockAuthUsecase{}, &mockUserUsecase{}),
		role.NewHandler(&mockRoleUsecase{}),
		permission.NewHandler(&mockPermissionUsecase{}),
		audit.NewHandler(&mockAuthUsecase{}),
		&mockAuthUsecase{},
		mockPermissionCheck,
		&mockJWKSProvider{},
		[]string{"*"},
		RateLimitConfig{LoginLimit: 5, LoginWindow: time.Minute, RefreshLimit: 20},
	)

	t.Run("register returns 201", func(t *testing.T) {
		body := `{"username":"newuser","email":"new@test.com","password":"Secure1pass"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}


type mockAuthUsecase struct{}

func (m *mockAuthUsecase) Register(ctx context.Context, params usecase.RegisterParams) (*usecase.AuthResult, error) {
	return &usecase.AuthResult{TokenPair: port.TokenPair{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 3600}}, nil
}
func (m *mockAuthUsecase) Login(ctx context.Context, params usecase.LoginParams) (*usecase.LoginResult, error) {
	return &usecase.LoginResult{TokenPair: port.TokenPair{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 3600}}, nil
}
func (m *mockAuthUsecase) RefreshToken(ctx context.Context, params usecase.RefreshTokenParams) (*usecase.LoginResult, error) {
	return &usecase.LoginResult{TokenPair: port.TokenPair{AccessToken: "at", RefreshToken: "rt", ExpiresIn: 3600}}, nil
}
func (m *mockAuthUsecase) Logout(ctx context.Context, params usecase.LogoutParams) error { return nil }
func (m *mockAuthUsecase) SwitchContext(ctx context.Context, sessionID string, userRoleID int64) (*usecase.AuthResult, error) {
	return &usecase.AuthResult{TokenPair: port.TokenPair{AccessToken: "at", ExpiresIn: 3600}}, nil
}
func (m *mockAuthUsecase) SetDefaultRole(ctx context.Context, userID, roleID int64) error { return nil }
func (m *mockAuthUsecase) Introspect(ctx context.Context, tokenString string) (*usecase.IntrospectResult, error) {
	return &usecase.IntrospectResult{Active: true, UserID: 1}, nil
}
func (m *mockAuthUsecase) ValidateToken(ctx context.Context, tokenString string) (*port.TokenClaims, error) {
	return &port.TokenClaims{UserID: 1, SessionID: "sess"}, nil
}
func (m *mockAuthUsecase) HasPermission(claims *port.TokenClaims, required domain.Permission) bool {
	return true
}
func (m *mockAuthUsecase) ListLoginAudits(ctx context.Context, userIDs []int64, p pagination.Params) ([]domain.LoginAudit, *pagination.Metadata, error) {
	return nil, &pagination.Metadata{}, nil
}

type mockUserUsecase struct{}

func (m *mockUserUsecase) ListUsers(ctx context.Context, p pagination.Params) ([]domain.User, *pagination.Metadata, error) {
	return nil, nil, nil
}
func (m *mockUserUsecase) GetUser(ctx context.Context, id int64) (*domain.User, error) {
	return &domain.User{ID: id, Username: "test"}, nil
}
func (m *mockUserUsecase) UpdateUser(ctx context.Context, params usecase.UpdateUserParams) (*domain.User, error) {
	return &domain.User{ID: params.ID}, nil
}
func (m *mockUserUsecase) DeleteUser(ctx context.Context, id int64) error { return nil }
func (m *mockUserUsecase) ListStaff(ctx context.Context, userID int64) ([]domain.UserRoleAssignment, error) {
	return nil, nil
}

type mockRoleUsecase struct{}

func (m *mockRoleUsecase) ListRoles(ctx context.Context, merchantID *int64) ([]domain.Role, error) {
	return nil, nil
}
func (m *mockRoleUsecase) CreateRole(ctx context.Context, params usecase.CreateRoleParams) (*domain.Role, error) {
	return &domain.Role{ID: 1, Name: params.Name}, nil
}
func (m *mockRoleUsecase) GetRole(ctx context.Context, id int64) (*domain.Role, error) {
	return &domain.Role{ID: id}, nil
}
func (m *mockRoleUsecase) UpdateRole(ctx context.Context, params usecase.UpdateRoleParams) (*domain.Role, error) {
	return &domain.Role{ID: params.ID}, nil
}
func (m *mockRoleUsecase) DeleteRole(ctx context.Context, id int64) error { return nil }
func (m *mockRoleUsecase) AssignRole(ctx context.Context, userID int64, roleName string) error {
	return nil
}
func (m *mockRoleUsecase) RemoveRole(ctx context.Context, userID int64, roleName string) error {
	return nil
}
func (m *mockRoleUsecase) AssignPermission(ctx context.Context, roleID int64, permission domain.Permission) error {
	return nil
}
func (m *mockRoleUsecase) RemovePermission(ctx context.Context, roleID int64, permission domain.Permission) error {
	return nil
}

type mockPermissionUsecase struct{}

func (m *mockPermissionUsecase) ListPermissions(ctx context.Context) ([]domain.Permission, error) {
	return nil, nil
}
func (m *mockPermissionUsecase) CreatePermission(ctx context.Context, permission domain.Permission) error {
	return nil
}
func (m *mockPermissionUsecase) DeletePermission(ctx context.Context, permission domain.Permission) error {
	return nil
}

type mockJWKSProvider struct{}

func (m *mockJWKSProvider) JWKS() port.JSONWebKeySet {
	return port.JSONWebKeySet{Keys: []port.JSONWebKey{}}
}

func mockPermissionCheck(claims *port.TokenClaims, required domain.Permission) bool { return true }
