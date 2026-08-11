package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/inventory/domain"
	"github.com/saleforge/pos/services/internal/inventory/port/repository"
)

type StockMovementUsecase interface {
	List(ctx context.Context, params ListMovementsParams) ([]domain.StockMovement, error)
}

type ListMovementsParams struct {
	MerchantID    int64
	BranchID      int64
	ProductItemID *int64
	From          *time.Time
	To            *time.Time
}

type stockMovementUsecase struct {
	repo repository.StockMovementRepository
}

func NewStockMovementUsecase(repo repository.StockMovementRepository) StockMovementUsecase {
	return &stockMovementUsecase{repo: repo}
}

func (uc *stockMovementUsecase) List(ctx context.Context, params ListMovementsParams) ([]domain.StockMovement, error) {
	if params.BranchID == 0 {
		return nil, domain.ErrInvalidStock
	}
	return uc.repo.List(ctx, params.MerchantID, params.BranchID, params.ProductItemID, params.From, params.To)
}

var _ StockMovementUsecase = (*stockMovementUsecase)(nil)
