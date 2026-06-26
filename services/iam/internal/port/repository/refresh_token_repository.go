package repository

import (
	"context"

	"github.com/saleforge/pos/services/iam/internal/domain"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	Revoke(ctx context.Context, id string) error
	RevokeByUser(ctx context.Context, userID string) error
}
