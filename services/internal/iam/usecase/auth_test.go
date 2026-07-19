package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
	"github.com/saleforge/pos/services/internal/iam/port/repository"
)

func TestAuthUsecase_Register(t *testing.T) {
	t.Parallel()

	validPerms := []domain.Permission{domain.UserRead}

	tests := []struct {
		name           string
		input          RegisterParams
		userRepo       repository.UserRepository
		roleRepo       repository.RoleRepository
		passwordHasher port.PasswordHasher
		tokenSigner    port.TokenSigner
		wantErr        error
	}{
		{
			name: "successful registration",
			input: RegisterParams{
				Username: "johndoe",
				Email:    "john@example.com",
				Password: "Securepass1",
				Roles:    []string{"viewer"},
			},
			userRepo: &mockUserRepo{},
			roleRepo: &mockRoleRepo{
				roles: map[int64]*domain.Role{
					1: {ID: 1, Name: "viewer", Permissions: validPerms},
				},
			},
			passwordHasher: &mockPasswordHasher{},
			tokenSigner:    &mockTokenSigner{signedToken: "token123"},
		},
		{
			name: "invalid role",
			input: RegisterParams{
				Username: "janedoe",
				Email:    "jane@example.com",
				Password: "Securepass1",
				Roles:    []string{"nonexistent_role"},
			},
			userRepo:       &mockUserRepo{},
			roleRepo:       &mockRoleRepo{},
			passwordHasher: &mockPasswordHasher{},
			tokenSigner:    &mockTokenSigner{},
			wantErr:        domain.ErrInvalidRole,
		},
		{
			name: "duplicate username",
			input: RegisterParams{
				Username: "existinguser",
				Email:    "new@example.com",
				Password: "Securepass1",
				Roles:    []string{"viewer"},
			},
			userRepo: &mockUserRepo{
				users: map[int64]*domain.User{
					1: {Username: "existinguser", Email: "old@example.com"},
				},
			},
			roleRepo:       &mockRoleRepo{},
			passwordHasher: &mockPasswordHasher{},
			tokenSigner:    &mockTokenSigner{},
			wantErr:        domain.ErrUserAlreadyExists,
		},
		{
			name: "duplicate email",
			input: RegisterParams{
				Username: "newuser",
				Email:    "dup@example.com",
				Password: "Securepass1",
				Roles:    []string{"viewer"},
			},
			userRepo: &mockUserRepo{
				users: map[int64]*domain.User{
					1: {Username: "other", Email: "dup@example.com"},
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

			uc := newTestUsecase(tt.userRepo, tt.roleRepo, tt.passwordHasher, tt.tokenSigner, &mockSessionStore{})
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
			if result.AccessToken == "" {
				t.Error("expected non-empty access token")
			}
			if result.RefreshToken == "" {
				t.Error("expected non-empty refresh token")
			}
		})
	}
}

func TestAuthUsecase_Login(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          LoginParams
		userRepo       repository.UserRepository
		roleRepo       repository.RoleRepository
		passwordHasher port.PasswordHasher
		tokenSigner    port.TokenSigner
		wantErr        error
	}{
		{
			name: "user not found",
			input: LoginParams{
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
			input: LoginParams{
				Username: "johndoe",
				Password: "securepass123",
			},
			userRepo: &mockUserRepo{
				users: map[int64]*domain.User{
					1: {ID: 1, Username: "johndoe", Password: "hashed:securepass123"},
				},
			},
			roleRepo: &mockRoleRepo{
				roles: map[int64]*domain.Role{
					1: {ID: 1, Name: "viewer", Permissions: []domain.Permission{domain.UserRead}},
				},
			},
			passwordHasher: &mockPasswordHasher{},
			tokenSigner:    &mockTokenSigner{signedToken: "token123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			uc := newTestUsecase(tt.userRepo, tt.roleRepo, tt.passwordHasher, tt.tokenSigner, &mockSessionStore{})
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
		wantUserID     int64
	}{
		{
			name:        "valid token",
			tokenString: "valid.jwt.token",
			userRepo: &mockUserRepo{
				users: map[int64]*domain.User{
					1: {ID: 1},
				},
			},
			passwordHasher: &mockPasswordHasher{},
			tokenSigner: &mockTokenSigner{
				claims: &port.TokenClaims{UserID: 1},
			},
			wantUserID: 1,
		},
		{
			name:           "user not found",
			tokenString:    "valid.jwt.token",
			userRepo:       &mockUserRepo{},
			passwordHasher: &mockPasswordHasher{},
			tokenSigner: &mockTokenSigner{
				claims: &port.TokenClaims{UserID: 999},
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

			sessionStore := &mockSessionStore{}
			uc := newTestUsecase(tt.userRepo, &mockRoleRepo{}, tt.passwordHasher, tt.tokenSigner, sessionStore)
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
				t.Errorf("expected user ID %d, got %d", tt.wantUserID, claims.UserID)
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

			uc := newTestUsecase(&mockUserRepo{}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})
			result := uc.HasPermission(tt.claims, tt.required)

			if result != tt.wantResult {
				t.Errorf("expected %v, got %v", tt.wantResult, result)
			}
		})
	}
}

func TestAuthUsecase_Logout(t *testing.T) {
	t.Parallel()

	sessionStore := &mockSessionStore{
		sessions: map[string]*domain.Session{
			"test-session": {ID: "test-session", UserID: 1},
		},
	}
	uc := newTestUsecase(&mockUserRepo{}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, sessionStore)

	err := uc.Logout(context.Background(), LogoutParams{SessionID: "test-session"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthUsecase_SwitchContext(t *testing.T) {
	t.Parallel()

	sessionStore := &mockSessionStore{
		sessions: map[string]*domain.Session{
			"test-session": {ID: "test-session", UserID: 1},
		},
	}
	userRepo := &mockUserRepo{
		users: map[int64]*domain.User{
			1: {
				ID: 1, Username: "test",
				Roles: []domain.UserRoleAssignment{
					{Role: domain.Role{ID: 2, Name: "manager"}, MerchantID: 100, IsDefault: true},
				},
			},
		},
	}
	uc := newTestUsecase(userRepo, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{signedToken: "new-access"}, sessionStore)

	result, err := uc.SwitchContext(context.Background(), "test-session", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if result.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
}

func TestAuthUsecase_SwitchContext_Forbidden(t *testing.T) {
	t.Parallel()

	sessionStore := &mockSessionStore{
		sessions: map[string]*domain.Session{
			"test-session": {ID: "test-session", UserID: 1},
		},
	}
	userRepo := &mockUserRepo{
		users: map[int64]*domain.User{
			1: {ID: 1, Username: "test", Roles: nil},
		},
	}
	uc := newTestUsecase(userRepo, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, sessionStore)

	_, err := uc.SwitchContext(context.Background(), "test-session", 99)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestAuthUsecase_SetDefaultRole(t *testing.T) {
	t.Parallel()

	uc := newTestUsecase(&mockUserRepo{}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})
	err := uc.SetDefaultRole(context.Background(), 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthUsecase_Introspect(t *testing.T) {
	t.Parallel()

	uc := newTestUsecase(&mockUserRepo{
		users: map[int64]*domain.User{
			1: {ID: 1, Username: "test", Status: domain.UserStatusActive},
		},
	}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{
		claims: &port.TokenClaims{UserID: 1, RoleName: "admin", Permissions: []domain.Permission{domain.UserRead}},
	}, &mockSessionStore{})

	result, err := uc.Introspect(context.Background(), "valid-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Active {
		t.Error("expected active introspection")
	}
	if result.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", result.UserID)
	}
}

func TestAuthUsecase_Introspect_InvalidToken(t *testing.T) {
	t.Parallel()

	uc := newTestUsecase(&mockUserRepo{}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{
		verifyErr: domain.ErrInvalidToken,
	}, &mockSessionStore{})

	result, err := uc.Introspect(context.Background(), "bad-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Active {
		t.Error("expected inactive for invalid token")
	}
}

func TestAuthUsecase_RefreshToken(t *testing.T) {
	t.Parallel()

	sessionStore := &mockSessionStore{
		sessions: map[string]*domain.Session{
			"test-session-id": {
				ID:               "test-session-id",
				UserID:           1,
				RefreshTokenHash: "hashed:refresh:token123",
				ExpiresAt:        time.Now().Add(30 * 24 * time.Hour),
			},
		},
	}
	userRepo := &mockUserRepo{
		users: map[int64]*domain.User{
			1: {ID: 1, Username: "test", Status: domain.UserStatusActive},
		},
	}
	uc := newTestUsecase(userRepo, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{
		signedToken:   "new-access-token",
		refreshUserID: 1,
	}, sessionStore)

	result, err := uc.RefreshToken(context.Background(), RefreshTokenParams{RefreshToken: "refresh:token123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken != "new-access-token" {
		t.Errorf("expected 'new-access-token', got %q", result.AccessToken)
	}
}

func TestAuthUsecase_RefreshToken_Invalid(t *testing.T) {
	t.Parallel()

	uc := newTestUsecase(&mockUserRepo{}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})
	_, err := uc.RefreshToken(context.Background(), RefreshTokenParams{RefreshToken: "bad"})
	if !errors.Is(err, domain.ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestAuthUsecase_Register_PasswordValidation(t *testing.T) {
	t.Parallel()

	uc := newTestUsecase(&mockUserRepo{}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})

	t.Run("password too short", func(t *testing.T) {
		_, err := uc.Register(context.Background(), RegisterParams{
			Username: "u", Email: "u@t.com", Password: "Ab1",
		})
		if !errors.Is(err, domain.ErrPasswordPolicy) {
			t.Errorf("expected password policy error, got %v", err)
		}
	})

	t.Run("password missing lowercase", func(t *testing.T) {
		_, err := uc.Register(context.Background(), RegisterParams{
			Username: "u", Email: "u@t.com", Password: "ABCDEF123",
		})
		if !errors.Is(err, domain.ErrPasswordPolicy) {
			t.Errorf("expected password policy error, got %v", err)
		}
	})

	t.Run("password missing uppercase", func(t *testing.T) {
		_, err := uc.Register(context.Background(), RegisterParams{
			Username: "u", Email: "u@t.com", Password: "abcdef123",
		})
		if !errors.Is(err, domain.ErrPasswordPolicy) {
			t.Errorf("expected password policy error, got %v", err)
		}
	})

	t.Run("password missing digit", func(t *testing.T) {
		_, err := uc.Register(context.Background(), RegisterParams{
			Username: "u", Email: "u@t.com", Password: "Abcdefgh",
		})
		if !errors.Is(err, domain.ErrPasswordPolicy) {
			t.Errorf("expected password policy error, got %v", err)
		}
	})
}
