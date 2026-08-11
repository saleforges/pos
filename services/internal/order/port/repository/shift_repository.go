package repository

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/order/domain"
)

type ShiftRepository interface {
	Create(ctx context.Context, shift *domain.Shift) error
	GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Shift, error)
	GetOpenByBranch(ctx context.Context, merchantID, branchID int64) (*domain.Shift, error)
	Update(ctx context.Context, shift *domain.Shift) error
	List(ctx context.Context, merchantID, branchID int64) ([]domain.Shift, error)
	SumCashPayments(ctx context.Context, merchantID, branchID int64, from, to time.Time) (float64, error)
}
