package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/iam/domain"
	"github.com/saleforge/pos/services/internal/iam/port"
)

// AuthUsecase defines authentication-related operations.
type AuthUsecase interface {
	Register(ctx context.Context, params RegisterParams) (*AuthResult, error)
	Login(ctx context.Context, params LoginParams) (*LoginResult, error)
	RefreshToken(ctx context.Context, params RefreshTokenParams) (*LoginResult, error)
	Logout(ctx context.Context, params LogoutParams) error
	SwitchContext(ctx context.Context, sessionID string, userRoleID int64) (*AuthResult, error)
	SetDefaultRole(ctx context.Context, userID, roleID int64) error
	Introspect(ctx context.Context, tokenString string) (*IntrospectResult, error)
	ValidateToken(ctx context.Context, tokenString string) (*port.TokenClaims, error)
	HasPermission(claims *port.TokenClaims, required domain.Permission) bool
}
