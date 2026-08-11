package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
	"github.com/saleforge/pos/services/pkg/pagination"
)

func TestAuthUsecase_Register(t *testing.T) {
	t.Parallel()

	roleRepo := &mockRoleRepo{}
	roleRepo.roles = map[int64]*domain.Role{
		1: {ID: 1, Name: "admin"},
	}
	tokenSigner := &mockTokenSigner{
		signedToken: "test-access-token",
	}

	t.Run("valid registration with default roles", func(t *testing.T) {
		uc := newTestAuthUsecase(&mockUserRepo{}, roleRepo, &mockPasswordHasher{}, tokenSigner, &mockSessionStore{})
		result, err := uc.Register(context.Background(), RegisterParams{
			Username:  "user1",
			Email:     "user1@test.com",
			Password:  "Secure1pass",
			Roles:     nil,
			UserType:  domain.UserTypeMerchant,
			IPAddress: "127.0.0.1",
			UserAgent: "test",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.AccessToken != "test-access-token" {
			t.Errorf("expected 'test-access-token', got %q", result.AccessToken)
		}
	})

	t.Run("password too short returns password policy error", func(t *testing.T) {
		uc := newTestAuthUsecase(&mockUserRepo{}, &mockRoleRepo{}, &mockPasswordHasher{}, tokenSigner, &mockSessionStore{})
		_, err := uc.Register(context.Background(), RegisterParams{
			Username:  "newuser",
			Email:     "newuser@t.com",
			Password:  "Ab1",
			UserType:  domain.UserTypeMerchant,
			IPAddress: "127.0.0.1",
			UserAgent: "test",
		})
		if !errors.Is(err, domain.ErrPasswordPolicy) {
			t.Errorf("expected password policy error, got %v", err)
		}
	})
}

func TestAuthUsecase_Login(t *testing.T) {
	t.Parallel()

	userRepo := &mockUserRepo{
		users: map[int64]*domain.User{
			1: {ID: 1, Username: "testuser", Password: "hashed:Correct1pass", Status: domain.UserStatusActive, Type: domain.UserTypeMerchant},
		},
	}
	roleRepo := &mockRoleRepo{}
	tokenSigner := &mockTokenSigner{
		signedToken: "test-access-token",
	}
	hasher := &mockPasswordHasher{}

	t.Run("valid credentials returns token", func(t *testing.T) {
		uc := newTestAuthUsecase(userRepo, roleRepo, hasher, tokenSigner, &mockSessionStore{})
		result, err := uc.Login(context.Background(), LoginParams{
			Username:  "testuser",
			Password:  "Correct1pass",
			IPAddress: "127.0.0.1",
			UserAgent: "test",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.AccessToken == "" {
			t.Error("expected non-empty access token")
		}
	})

	t.Run("invalid password returns error", func(t *testing.T) {
		uc := newTestAuthUsecase(userRepo, roleRepo, hasher, tokenSigner, &mockSessionStore{})
		_, err := uc.Login(context.Background(), LoginParams{
			Username:  "testuser",
			Password:  "WrongPassword1",
			IPAddress: "127.0.0.1",
			UserAgent: "test",
		})
		if !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})
}

func TestAuthUsecase_ValidateToken(t *testing.T) {
	t.Parallel()

	userRepo := &mockUserRepo{
		users: map[int64]*domain.User{
			1: {ID: 1, Username: "testuser", Status: domain.UserStatusActive},
		},
	}
	tokenSigner := &mockTokenSigner{
		claims: &port.TokenClaims{UserID: 1, SessionID: "sess-1"},
	}
	sessionStore := &mockSessionStore{
		sessions: map[string]*domain.Session{
			"sess-1": {ID: "sess-1", UserID: 1, ExpiresAt: time.Now().Add(30 * 24 * time.Hour)},
		},
	}

	t.Run("valid token returns claims", func(t *testing.T) {
		uc := newTestAuthUsecase(userRepo, &mockRoleRepo{}, &mockPasswordHasher{}, tokenSigner, sessionStore)
		claims, err := uc.ValidateToken(context.Background(), "valid-token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if claims.UserID != 1 {
			t.Errorf("expected UserID 1, got %d", claims.UserID)
		}
	})

	t.Run("invalid token returns error", func(t *testing.T) {
		signer := &mockTokenSigner{verifyErr: domain.ErrInvalidToken}
		uc := newTestAuthUsecase(userRepo, &mockRoleRepo{}, &mockPasswordHasher{}, signer, sessionStore)
		_, err := uc.ValidateToken(context.Background(), "bad-token")
		if !errors.Is(err, domain.ErrInvalidToken) {
			t.Errorf("expected ErrInvalidToken, got %v", err)
		}
	})
}

func TestAuthUsecase_Introspect(t *testing.T) {
	t.Parallel()

	t.Run("active user returns active result", func(t *testing.T) {
		userRepo := &mockUserRepo{
			users: map[int64]*domain.User{
				1: {ID: 1, Username: "test", Status: domain.UserStatusActive},
			},
		}
		uc := newTestAuthUsecase(userRepo, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{
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
	})

	t.Run("invalid token returns inactive", func(t *testing.T) {
		uc := newTestAuthUsecase(&mockUserRepo{}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{
			verifyErr: domain.ErrInvalidToken,
		}, &mockSessionStore{})

		result, err := uc.Introspect(context.Background(), "bad-token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Active {
			t.Error("expected inactive for invalid token")
		}
	})

	t.Run("user not found returns inactive", func(t *testing.T) {
		userRepo := &mockUserRepo{} // empty — no users
		uc := newTestAuthUsecase(userRepo, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{
			claims: &port.TokenClaims{UserID: 999, RoleName: "admin"},
		}, &mockSessionStore{})

		result, err := uc.Introspect(context.Background(), "valid-token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Active {
			t.Error("expected inactive for non-existent user")
		}
	})

	t.Run("DB error returns internal error", func(t *testing.T) {
		userRepo := &mockUserRepo{err: errors.New("connection refused")} // simulate DB down
		uc := newTestAuthUsecase(userRepo, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{
			claims: &port.TokenClaims{UserID: 1, RoleName: "admin"},
		}, &mockSessionStore{})

		_, err := uc.Introspect(context.Background(), "valid-token")
		if !errors.Is(err, domain.ErrInternal) {
			t.Errorf("expected ErrInternal for DB error, got %v", err)
		}
	})

	t.Run("disabled user returns inactive", func(t *testing.T) {
		userRepo := &mockUserRepo{
			users: map[int64]*domain.User{
				1: {ID: 1, Username: "test", Status: domain.UserStatusDisabled},
			},
		}
		uc := newTestAuthUsecase(userRepo, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{
			claims: &port.TokenClaims{UserID: 1, RoleName: "admin"},
		}, &mockSessionStore{})

		result, err := uc.Introspect(context.Background(), "valid-token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Active {
			t.Error("expected inactive for disabled user")
		}
	})

	t.Run("staff repo error returns internal error", func(t *testing.T) {
		t.Skip("staff repo error path covered by handler test")
	})
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
	uc := newTestAuthUsecase(userRepo, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{
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

	uc := newTestAuthUsecase(&mockUserRepo{}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})
	_, err := uc.RefreshToken(context.Background(), RefreshTokenParams{RefreshToken: "bad"})
	if !errors.Is(err, domain.ErrInvalidRefreshToken) {
		t.Errorf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestAuthUsecase_Register_PasswordValidation(t *testing.T) {
	t.Parallel()

	uc := newTestAuthUsecase(&mockUserRepo{}, &mockRoleRepo{}, &mockPasswordHasher{}, &mockTokenSigner{}, &mockSessionStore{})

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

func TestAuthUsecase_ListLoginAudits(t *testing.T) {
	t.Parallel()

	loginAuditRepo := &mockLoginAuditRepo{
		listResult: []domain.LoginAudit{
			{ID: 1, UserID: 5, Email: "cashier1@test.com", Success: true},
		},
	}
	uc := NewAuthUsecase(
		&mockUserRepo{}, &mockRoleRepo{}, &mockPermissionRepo{}, loginAuditRepo,
		&mockStaffRepo{}, &mockSessionStore{}, &mockEventPublisher{},
		&mockPasswordHasher{}, &mockTokenSigner{}, &mockTokenHasher{}, nil, nil,
	)

	result, meta, err := uc.ListLoginAudits(context.Background(), []int64{5}, pagination.Params{Offset: 0, Limit: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].UserID != 5 {
		t.Errorf("unexpected result: %+v", result)
	}
	if meta.Count != 1 {
		t.Errorf("expected count 1, got %d", meta.Count)
	}
	if len(loginAuditRepo.listUserIDs) != 1 || loginAuditRepo.listUserIDs[0] != 5 {
		t.Errorf("expected repo to be called with userIDs [5], got %v", loginAuditRepo.listUserIDs)
	}
}
