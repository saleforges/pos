package port

import (
	"context"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

type SessionStore interface {
	Create(ctx context.Context, session *domain.Session) error
	Get(ctx context.Context, id string) (*domain.Session, error)
	Update(ctx context.Context, session *domain.Session) error
	Delete(ctx context.Context, id string) error
}
