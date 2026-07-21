package port

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

type UserCache interface {
	Get(ctx context.Context, id int64) (*domain.User, error)
	Set(ctx context.Context, u *domain.User, ttl time.Duration) error
	Delete(ctx context.Context, id int64) error
	Close() error
}
