package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	Revoke(ctx context.Context, id int64) error
	RevokeByUser(ctx context.Context, userID int64) error
}
