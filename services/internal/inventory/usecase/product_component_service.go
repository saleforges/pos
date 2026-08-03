package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/inventory/domain"
	"github.com/saleforge/pos/services/internal/inventory/port/repository"
)

type productComponentUsecase struct {
	repo repository.ProductComponentRepository
}

func NewProductComponentUsecase(repo repository.ProductComponentRepository) ProductComponentUsecase {
	return &productComponentUsecase{repo: repo}
}

func (uc *productComponentUsecase) Create(ctx context.Context, params CreateProductComponentParams) (*domain.ProductComponent, error) {
	now := time.Now().UTC()

	items := make([]domain.ProductComponentItem, len(params.Items))
	for i, p := range params.Items {
		items[i] = domain.ProductComponentItem{
			ComponentProductItemID: p.ComponentProductItemID,
			Quantity:               p.Quantity,
			UnitID:                 p.UnitID,
			CreatedAt:              now,
		}
	}

	component := &domain.ProductComponent{
		MerchantID:    params.MerchantID,
		ProductItemID: params.ProductItemID,
		CreatedAt:     now,
		UpdatedAt:     now,
		Items:         items,
	}

	if err := component.Validate(); err != nil {
		return nil, err
	}

	if err := uc.repo.Create(ctx, component); err != nil {
		return nil, err
	}
	return component, nil
}

func (uc *productComponentUsecase) GetByProductItem(ctx context.Context, productItemID int64, merchantID int64) (*domain.ProductComponent, error) {
	return uc.repo.GetByProductItem(ctx, productItemID, merchantID)
}

func (uc *productComponentUsecase) List(ctx context.Context, merchantID int64) ([]domain.ProductComponent, error) {
	return uc.repo.List(ctx, merchantID)
}

func (uc *productComponentUsecase) Update(ctx context.Context, params UpdateProductComponentParams) (*domain.ProductComponent, error) {
	// Get existing component
	existing, err := uc.repo.GetByProductItem(ctx, params.ProductItemID, params.MerchantID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	items := make([]domain.ProductComponentItem, len(params.Items))
	for i, p := range params.Items {
		items[i] = domain.ProductComponentItem{
			ComponentProductItemID: p.ComponentProductItemID,
			Quantity:               p.Quantity,
			UnitID:                 p.UnitID,
			CreatedAt:              now,
		}
	}

	existing.Items = items
	existing.UpdatedAt = now

	if err := existing.Validate(); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return existing, nil
}

func (uc *productComponentUsecase) Delete(ctx context.Context, id int64, merchantID int64) error {
	return uc.repo.Delete(ctx, id, merchantID)
}

var _ ProductComponentUsecase = (*productComponentUsecase)(nil)
