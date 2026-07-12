package port

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

type UserCache interface {
	Get(ctx context.Context, id int64) (*domain.User, bool)
	Set(ctx context.Context, u *domain.User, ttl time.Duration)
	Delete(ctx context.Context, id int64)
	Close() error
}
