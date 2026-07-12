package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

type StaffRepository interface {
	ListByUserID(ctx context.Context, userID int64) ([]domain.UserRoleAssignment, error)
	Create(ctx context.Context, userID int64, merchantID int64, merchantName, role string) error
}
