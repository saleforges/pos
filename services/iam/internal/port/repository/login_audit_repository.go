package repository

import (
	"context"

	"github.com/saleforge/pos/services/iam/internal/domain"
)

type LoginAuditRepository interface {
	Create(ctx context.Context, audit *domain.LoginAudit) error
	List(ctx context.Context, offset, limit int) ([]domain.LoginAudit, error)
}
