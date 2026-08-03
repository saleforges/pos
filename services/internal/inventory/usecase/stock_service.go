package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/inventory/domain"
	"github.com/saleforge/pos/services/internal/inventory/port/repository"
)

type stockUsecase struct {
	repo repository.StockRepository
}

func NewStockUsecase(repo repository.StockRepository) StockUsecase {
	return &stockUsecase{repo: repo}
}

func (uc *stockUsecase) Create(ctx context.Context, params CreateStockParams) (*domain.Stock, error) {
	now := time.Now().UTC()
	stock := &domain.Stock{
		MerchantID:    params.MerchantID,
		BranchID:      params.BranchID,
		ProductItemID: params.ProductItemID,
		Available:     params.Available,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := stock.Validate(); err != nil {
		return nil, err
	}

	if err := uc.repo.Create(ctx, stock); err != nil {
		return nil, err
	}
	return stock, nil
}

func (uc *stockUsecase) GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Stock, error) {
	return uc.repo.GetByID(ctx, id, merchantID)
}

func (uc *stockUsecase) List(ctx context.Context, merchantID int64) ([]domain.Stock, error) {
	return uc.repo.List(ctx, merchantID)
}

func (uc *stockUsecase) Update(ctx context.Context, params UpdateStockParams) (*domain.Stock, error) {
	stock, err := uc.repo.GetByID(ctx, params.ID, params.MerchantID)
	if err != nil {
		return nil, err
	}

	stock.Available = params.Available
	if err := stock.Validate(); err != nil {
		return nil, err
	}

	stock.UpdatedAt = time.Now().UTC()

	if err := uc.repo.Update(ctx, stock); err != nil {
		return nil, err
	}
	return stock, nil
}

var _ StockUsecase = (*stockUsecase)(nil)
