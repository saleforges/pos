package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/order/domain"
)

type ShiftUsecase interface {
	Open(ctx context.Context, params OpenShiftParams) (*domain.Shift, error)
	Close(ctx context.Context, params CloseShiftParams) (*domain.Shift, error)
	GetActive(ctx context.Context, merchantID, branchID int64) (*domain.Shift, error)
	List(ctx context.Context, merchantID, branchID int64) ([]domain.Shift, error)
}

type OpenShiftParams struct {
	MerchantID   int64
	BranchID     int64
	OpenedBy     int64
	StartingCash float64
}

type CloseShiftParams struct {
	ShiftID    int64
	MerchantID int64
	ClosedBy   int64
	ActualCash float64
	Note       string
}
