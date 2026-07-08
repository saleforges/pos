package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/iam/domain"
)

type StaffRepository interface {
	ListByUserID(ctx context.Context, userID string) ([]domain.StaffInfo, error)
	Create(ctx context.Context, userID, merchantID, merchantName, role string) error
}
