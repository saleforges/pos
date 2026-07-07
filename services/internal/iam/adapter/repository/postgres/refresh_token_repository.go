package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/saleforge/pos/services/pkg/otel"
	"github.com/saleforge/pos/services/internal/iam/domain"
)

type RefreshTokenRepository struct {
	pool *otel.TracedPool
}

func NewRefreshTokenRepository(pool *otel.TracedPool) *RefreshTokenRepository {
	return &RefreshTokenRepository{pool: pool}
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (r *RefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at) VALUES ($1, $2, $3, $4, $5)`,
		token.ID, token.UserID, hashToken(token.Token), token.ExpiresAt, token.CreatedAt,
	)
	return err
}

func (r *RefreshTokenRepository) GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	var rt domain.RefreshToken
	var revokedAt *time.Time
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, created_at, revoked_at FROM refresh_tokens WHERE token_hash = $1`,
		hashToken(token),
	).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.CreatedAt, &revokedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrInvalidRefreshToken
		}
		return nil, err
	}
	rt.RevokedAt = revokedAt
	rt.Token = token
	return &rt, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id string) error {
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = $1 WHERE id = $2`, now, id)
	return err
}

func (r *RefreshTokenRepository) RevokeByUser(ctx context.Context, userID string) error {
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx,
		`UPDATE refresh_tokens SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL`,
		now, userID,
	)
	return err
}
