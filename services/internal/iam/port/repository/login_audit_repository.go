package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

type LoginAuditRepository interface {
	Create(ctx context.Context, audit *domain.LoginAudit) error
	List(ctx context.Context, userIDs []int64, offset, limit int) ([]domain.LoginAudit, int64, error)
}
