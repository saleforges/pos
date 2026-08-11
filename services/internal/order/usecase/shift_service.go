package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/order/domain"
	"github.com/saleforge/pos/services/internal/order/port/repository"
)

type shiftUsecase struct {
	repo repository.ShiftRepository
}

func NewShiftUsecase(repo repository.ShiftRepository) ShiftUsecase {
	return &shiftUsecase{repo: repo}
}

func (uc *shiftUsecase) Open(ctx context.Context, params OpenShiftParams) (*domain.Shift, error) {
	_, err := uc.repo.GetOpenByBranch(ctx, params.MerchantID, params.BranchID)
	if err == nil {
		return nil, domain.ErrShiftAlreadyOpen
	}
	if err != domain.ErrShiftNotFound {
		return nil, err
	}

	shift := &domain.Shift{
		MerchantID:   params.MerchantID,
		BranchID:     params.BranchID,
		OpenedBy:     params.OpenedBy,
		Status:       domain.ShiftStatusOpen,
		StartingCash: params.StartingCash,
		OpenedAt:     time.Now().UTC(),
	}
	if err := shift.Validate(); err != nil {
		return nil, err
	}
	if err := uc.repo.Create(ctx, shift); err != nil {
		return nil, err
	}
	return shift, nil
}

func (uc *shiftUsecase) Close(ctx context.Context, params CloseShiftParams) (*domain.Shift, error) {
	shift, err := uc.repo.GetByID(ctx, params.ShiftID, params.MerchantID)
	if err != nil {
		return nil, err
	}
	if shift.Status == domain.ShiftStatusClosed {
		return nil, domain.ErrShiftAlreadyClosed
	}
	if params.ActualCash < 0 {
		return nil, domain.ErrInvalidShift
	}

	now := time.Now().UTC()
	cashSales, err := uc.repo.SumCashPayments(ctx, params.MerchantID, shift.BranchID, shift.OpenedAt, now)
	if err != nil {
		return nil, err
	}
	expected := shift.StartingCash + cashSales
	actualCash := params.ActualCash
	variance := actualCash - expected
	closedBy := params.ClosedBy

	shift.Status = domain.ShiftStatusClosed
	shift.ClosedBy = &closedBy
	shift.ExpectedCash = &expected
	shift.ActualCash = &actualCash
	shift.Variance = &variance
	shift.Note = params.Note
	shift.ClosedAt = &now

	if err := uc.repo.Update(ctx, shift); err != nil {
		return nil, err
	}
	return shift, nil
}

func (uc *shiftUsecase) GetActive(ctx context.Context, merchantID, branchID int64) (*domain.Shift, error) {
	return uc.repo.GetOpenByBranch(ctx, merchantID, branchID)
}

func (uc *shiftUsecase) List(ctx context.Context, merchantID, branchID int64) ([]domain.Shift, error) {
	return uc.repo.List(ctx, merchantID, branchID)
}

var _ ShiftUsecase = (*shiftUsecase)(nil)
