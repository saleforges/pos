package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/saleforge/pos/services/iam/internal/domain"
	"github.com/saleforge/pos/services/iam/internal/port"
	"github.com/saleforge/pos/services/iam/internal/port/repository"
)

type mockUserRepo struct {
	users map[string]*domain.User
	err   error
}

func (m *mockUserRepo) Create(_ context.Context, user *domain.User) error {
	if m.err != nil {
		return m.err
	}
	if m.users == nil {
		m.users = make(map[string]*domain.User)
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetByID(_ context.Context, id string) (*domain.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	u, ok := m.users[id]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) GetByUsername(_ context.Context, username string) (*domain.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, domain.ErrUserNotFound
}

func (m *mockUserRepo) List(_ context.Context, offset, limit int) ([]domain.User, error) {
	return nil, nil
}

func (m *mockUserRepo) Update(_ context.Context, user *domain.User) error {
	if m.err != nil {
		return m.err
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) Delete(_ context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.users, id)
	return nil
}

func (m *mockUserRepo) AddRole(_ context.Context, userID, roleName string) error {
	return nil
}

func (m *mockUserRepo) RemoveRole(_ context.Context, userID, roleName string) error {
	return nil
}

type mockRoleRepo struct {
	roles       map[string]*domain.Role
	permissions map[string][]domain.Permission
	err         error
}

func (m *mockRoleRepo) Create(_ context.Context, role *domain.Role) error {
	if m.err != nil {
		return m.err
	}
	if m.roles == nil {
		m.roles = make(map[string]*domain.Role)
	}
	m.roles[role.Name] = role
	return nil
}

func (m *mockRoleRepo) GetByName(_ context.Context, name string) (*domain.Role, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.roles != nil {
		if r, ok := m.roles[name]; ok {
			return r, nil
		}
	}
	return nil, domain.ErrInvalidRole
}

func (m *mockRoleRepo) List(_ context.Context) ([]domain.Role, error) {
	return nil, nil
}

func (m *mockRoleRepo) Update(_ context.Context, role *domain.Role) error {
	if m.err != nil {
		return m.err
	}
	if m.roles != nil {
		m.roles[role.Name] = role
	}
	return nil
}

func (m *mockRoleRepo) Delete(_ context.Context, name string) error {
	return nil
}

func (m *mockRoleRepo) AddPermission(_ context.Context, roleName string, permission domain.Permission) error {
	return nil
}

func (m *mockRoleRepo) RemovePermission(_ context.Context, roleName string, permission domain.Permission) error {
	return nil
}

func (m *mockRoleRepo) GetPermissions(_ context.Context, roleName string) ([]domain.Permission, error) {
	if m.err != nil {
		return nil, m.err
	}
	perms, ok := m.permissions[roleName]
	if !ok {
		return nil, domain.ErrInvalidRole
	}
	return perms, nil
}

type mockPermissionRepo struct {
	permissions map[domain.Permission]bool
	err         error
}

func (m *mockPermissionRepo) Create(_ context.Context, p domain.Permission) error {
	return nil
}

func (m *mockPermissionRepo) GetAll(_ context.Context) ([]domain.Permission, error) {
	result := make([]domain.Permission, 0, len(m.permissions))
	for p := range m.permissions {
		result = append(result, p)
	}
	return result, nil
}

func (m *mockPermissionRepo) Delete(_ context.Context, p domain.Permission) error {
	return nil
}

type mockRefreshTokenRepo struct {
	tokens map[string]*domain.RefreshToken
	err    error
}

func (m *mockRefreshTokenRepo) Create(_ context.Context, token *domain.RefreshToken) error {
	if m.tokens == nil {
		m.tokens = make(map[string]*domain.RefreshToken)
	}
	m.tokens[token.ID] = token
	return nil
}

func (m *mockRefreshTokenRepo) GetByToken(_ context.Context, token string) (*domain.RefreshToken, error) {
	return nil, domain.ErrInvalidRefreshToken
}

func (m *mockRefreshTokenRepo) Revoke(_ context.Context, id string) error {
	return nil
}

func (m *mockRefreshTokenRepo) RevokeByUser(_ context.Context, userID string) error {
	return nil
}

type mockSessionRepo struct {
	err error
}

func (m *mockSessionRepo) Create(_ context.Context, session *domain.Session) error {
	return nil
}

func (m *mockSessionRepo) GetByID(_ context.Context, id string) (*domain.Session, error) {
	return nil, nil
}

func (m *mockSessionRepo) ListByUser(_ context.Context, userID string) ([]domain.Session, error) {
	return nil, nil
}

func (m *mockSessionRepo) ListAll(_ context.Context) ([]domain.Session, error) {
	return nil, nil
}

func (m *mockSessionRepo) Revoke(_ context.Context, id string) error {
	return nil
}

func (m *mockSessionRepo) RevokeAll(_ context.Context) error {
	return nil
}

func (m *mockSessionRepo) RevokeByUser(_ context.Context, userID string) error {
	return nil
}

type mockAPIKeyRepo struct {
	err error
}

func (m *mockAPIKeyRepo) Create(_ context.Context, key *domain.APIKey) error {
	return nil
}

func (m *mockAPIKeyRepo) GetByID(_ context.Context, id string) (*domain.APIKey, error) {
	return nil, nil
}

func (m *mockAPIKeyRepo) GetByKey(_ context.Context, key string) (*domain.APIKey, error) {
	return nil, nil
}

func (m *mockAPIKeyRepo) ListByUser(_ context.Context, userID string) ([]domain.APIKey, error) {
	return nil, nil
}

func (m *mockAPIKeyRepo) ListAll(_ context.Context) ([]domain.APIKey, error) {
	return nil, nil
}

func (m *mockAPIKeyRepo) Revoke(_ context.Context, id string) error {
	return nil
}

type mockLoginAuditRepo struct {
	err error
}

func (m *mockLoginAuditRepo) Create(_ context.Context, audit *domain.LoginAudit) error {
	return nil
}

func (m *mockLoginAuditRepo) List(_ context.Context, offset, limit int) ([]domain.LoginAudit, error) {
	return nil, nil
}

type mockEventPublisher struct {
	err error
}

func (m *mockEventPublisher) Publish(_ context.Context, _ string, _ interface{}) error {
	return nil
}

type mockPasswordHasher struct {
	hashErr    error
	compareErr error
}

func (m *mockPasswordHasher) Hash(password string) (string, error) {
	if m.hashErr != nil {
		return "", m.hashErr
	}
	return "hashed:" + password, nil
}

func (m *mockPasswordHasher) Compare(hashedPassword, password string) error {
	if m.compareErr != nil {
		return m.compareErr
	}
	if hashedPassword != "hashed:"+password {
		return domain.ErrInvalidCredentials
	}
	return nil
}

type mockTokenSigner struct {
	signedToken      string
	claims           *port.TokenClaims
	signErr          error
	verifyErr        error
	refreshUserID    string
	refreshSignErr   error
	refreshVerifyErr error
}

func (m *mockTokenSigner) SignAccessToken(_ port.TokenClaims) (string, error) {
	if m.signErr != nil {
		return "", m.signErr
	}
	return m.signedToken, nil
}

func (m *mockTokenSigner) SignRefreshToken(_ string) (string, error) {
	if m.refreshSignErr != nil {
		return "", m.refreshSignErr
	}
	return "refresh:" + m.signedToken, nil
}

func (m *mockTokenSigner) VerifyAccessToken(_ string) (*port.TokenClaims, error) {
	if m.verifyErr != nil {
		return nil, m.verifyErr
	}
	return m.claims, nil
}

func (m *mockTokenSigner) VerifyRefreshToken(_ string) (string, error) {
	if m.refreshVerifyErr != nil {
		return "", m.refreshVerifyErr
	}
	return m.refreshUserID, nil
}

func newTestUsecase(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	passwordHasher port.PasswordHasher,
	tokenSigner port.TokenSigner,
) *AuthUsecase {
	return NewAuthUsecase(
		userRepo,
		roleRepo,
		&mockPermissionRepo{},
		&mockRefreshTokenRepo{},
		&mockLoginAuditRepo{},
		&mockEventPublisher{},
		passwordHasher,
		tokenSigner,
	)
}

func TestAuthUsecase_Register(t *testing.T) {
	t.Parallel()

	validPerms := []domain.Permission{domain.UserRead}

	tests := []struct {
		name           string
		input          RegisterInput
		userRepo       repository.UserRepository
		roleRepo       repository.RoleRepository
		passwordHasher port.PasswordHasher
		tokenSigner    port.TokenSigner
		wantErr        error
	}{
		{
			name: "successful registration",
			input: RegisterInput{
				Username: "johndoe",
				Email:    "john@example.com",
				Password: "Securepass1",
				Roles:    []string{"viewer"},
			},
			userRepo:       &mockUserRepo{},
			roleRepo:       &mockRoleRepo{permissions: map[string][]domain.Permission{"viewer": validPerms}},
			passwordHasher: &mockPasswordHasher{},
			tokenSigner:    &mockTokenSigner{signedToken: "token123"},
		},
		{
			name: "invalid role",
			input: RegisterInput{
				Username: "janedoe",
				Email:    "jane@example.com",
				Password: "Securepass1",
				Roles:    []string{"superadmin"},
			},
			userRepo:       &mockUserRepo{},
			roleRepo:       &mockRoleRepo{},
			passwordHasher: &mockPasswordHasher{},
			tokenSigner:    &mockTokenSigner{},
			wantErr:        domain.ErrInvalidRole,
		},
		{
			name: "duplicate username",
			input: RegisterInput{
				Username: "existinguser",
				Email:    "new@example.com",
				Password: "Securepass1",
				Roles:    []string{"viewer"},
			},
			userRepo: &mockUserRepo{
				users: map[string]*domain.User{
					"1": {Username: "existinguser", Email: "old@example.com"},
				},
			},
			roleRepo:       &mockRoleRepo{},
			passwordHasher: &mockPasswordHasher{},
			tokenSigner:    &mockTokenSigner{},
			wantErr:        domain.ErrUserAlreadyExists,
		},
		{
			name: "duplicate email",
			input: RegisterInput{
				Username: "newuser",
				Email:    "dup@example.com",
				Password: "Securepass1",
				Roles:    []string{"viewer"},
			},
			userRepo: &mockUserRepo{
				users: map[string]*domain.User{
					"1": {Username: "other", Email: "dup@example.com"},
				},
			},
			roleRepo:       &mockRoleRepo{},
			passwordHasher: &mockPasswordHasher{},
			tokenSigner:    &mockTokenSigner{},
			wantErr:        domain.ErrUserAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := newTestUsecase(tt.userRepo, tt.roleRepo, tt.passwordHasher, tt.tokenSigner)
			result, err := uc.Register(context.Background(), tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result == nil {
				t.Fatal("expected result, got nil")
			}
			if result.Token == "" {
				t.Error("expected non-empty token")
			}
			if result.User.Username != tt.input.Username {
				t.Errorf("expected username %s, got %s", tt.input.Username, result.User.Username)
			}
			if result.User.Password == "" {
				t.Error("password hash should be set on the domain object")
			}
		})
	}
}

func TestAuthUsecase_Login(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          LoginInput
		userRepo       repository.UserRepository
		roleRepo       repository.RoleRepository
		passwordHasher port.PasswordHasher
		tokenSigner    port.TokenSigner
		wantErr        error
	}{
		{
			name: "user not found",
			input: LoginInput{
				Username: "nonexistent",
				Password: "anypass",
			},
			userRepo:       &mockUserRepo{},
			roleRepo:       &mockRoleRepo{},
			passwordHasher: &mockPasswordHasher{},
			tokenSigner:    &mockTokenSigner{},
			wantErr:        domain.ErrInvalidCredentials,
		},
		{
			name: "successful login",
			input: LoginInput{
				Username: "johndoe",
				Password: "securepass123",
			},
			userRepo: &mockUserRepo{
				users: map[string]*domain.User{
					"u1": {ID: "u1", Username: "johndoe", Password: "hashed:securepass123"},
				},
			},
			roleRepo:       &mockRoleRepo{permissions: map[string][]domain.Permission{"viewer": {domain.UserRead}}},
			passwordHasher: &mockPasswordHasher{},
			tokenSigner:    &mockTokenSigner{signedToken: "token123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := newTestUsecase(tt.userRepo, tt.roleRepo, tt.passwordHasher, tt.tokenSigner)
			_, err := uc.Login(context.Background(), tt.input)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestAuthUsecase_ValidateToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		tokenString    string
		userRepo       repository.UserRepository
		passwordHasher port.PasswordHasher
		tokenSigner    port.TokenSigner
		wantErr        error
		wantUserID     string
	}{
		{
			name:        "valid token",
			tokenString: "valid.jwt.token",
			userRepo: &mockUserRepo{
				users: map[string]*domain.User{
					"user-1": {ID: "user-1"},
				},
			},
			passwordHasher: &mockPasswordHasher{},
			tokenSigner: &mockTokenSigner{
				claims: &port.TokenClaims{UserID: "user-1"},
			},
			wantUserID: "user-1",
		},
		{
			name:           "user not found",
			tokenString:    "valid.jwt.token",
			userRepo:       &mockUserRepo{},
			passwordHasher: &mockPasswordHasher{},
			tokenSigner: &mockTokenSigner{
				claims: &port.TokenClaims{UserID: "missing-user"},
			},
			wantErr: domain.ErrInvalidToken,
		},
		{
			name:           "token verification fails",
			tokenString:    "invalid.jwt.token",
			userRepo:       &mockUserRepo{},
			passwordHasher: &mockPasswordHasher{},
			tokenSigner:    &mockTokenSigner{verifyErr: domain.ErrInvalidToken},
			wantErr:        domain.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := newTestUsecase(tt.userRepo, &mockRoleRepo{}, tt.passwordHasher, tt.tokenSigner)
			claims, err := uc.ValidateToken(context.Background(), tt.tokenString)

			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if claims.UserID != tt.wantUserID {
				t.Errorf("expected user ID %s, got %s", tt.wantUserID, claims.UserID)
			}
		})
	}
}

func TestAuthUsecase_HasPermission(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		claims     *port.TokenClaims
		required   domain.Permission
		wantResult bool
	}{
		{
			name: "has permission",
			claims: &port.TokenClaims{
				Permissions: []domain.Permission{domain.UserRead, domain.UserCreate},
			},
			required:   domain.UserRead,
			wantResult: true,
		},
		{
			name: "missing permission",
			claims: &port.TokenClaims{
				Permissions: []domain.Permission{domain.UserRead},
			},
			required:   domain.UserDelete,
			wantResult: false,
		},
		{
			name: "empty permissions",
			claims: &port.TokenClaims{
				Permissions: []domain.Permission{},
			},
			required:   domain.UserRead,
			wantResult: false,
		},
		{
			name:       "nil permissions",
			claims:     &port.TokenClaims{},
			required:   domain.UserRead,
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := newTestUsecase(&mockUserRepo{}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{})
			result := uc.HasPermission(tt.claims, tt.required)

			if result != tt.wantResult {
				t.Errorf("expected %v, got %v", tt.wantResult, result)
			}
		})
	}
}
