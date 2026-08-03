package usecase

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/saleforge/pos/services/internal/inventory/domain"
	"github.com/saleforge/pos/services/internal/inventory/port/repository"
)

type stockUsecase struct {
	repo           repository.StockRepository
	adjustmentRepo repository.StockAdjustmentRepository
	componentRepo  repository.ProductComponentRepository
}

func NewStockUsecase(repo repository.StockRepository, adjustmentRepo repository.StockAdjustmentRepository, componentRepo repository.ProductComponentRepository) StockUsecase {
	return &stockUsecase{repo: repo, adjustmentRepo: adjustmentRepo, componentRepo: componentRepo}
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

func (uc *stockUsecase) Deduct(ctx context.Context, params AdjustStockParams) error {
	if params.BranchID == 0 || len(params.Items) == 0 {
		return domain.ErrInvalidStock
	}
	items, err := uc.expandComponents(ctx, params.MerchantID, params.Items)
	if err != nil {
		return err
	}
	return uc.adjustmentRepo.Deduct(ctx, params.MerchantID, params.BranchID, params.ReferenceType, params.ReferenceID, items)
}

func (uc *stockUsecase) Restore(ctx context.Context, params AdjustStockParams) error {
	if params.BranchID == 0 || len(params.Items) == 0 {
		return domain.ErrInvalidStock
	}
	items, err := uc.expandComponents(ctx, params.MerchantID, params.Items)
	if err != nil {
		return err
	}
	return uc.adjustmentRepo.Restore(ctx, params.MerchantID, params.BranchID, params.ReferenceType, params.ReferenceID, items)
}

// expandComponents resolves product components for every sold item and
// merges component raw-material quantities so the adjustment batch is flat.
// E.g. selling 2x Es Teh (component: 1x Gula each) produces a batch of
// [EsTeh:2, Gula:2]. Each component quantity is converted to the raw
// material's stock unit via its conversion factor BEFORE merging, so items
// sharing a raw material in different units (0.5 kg + 2 packs) sum
// correctly. Fractional results are rounded up to never undersell.
func (uc *stockUsecase) expandComponents(ctx context.Context, merchantID int64, items []AdjustStockItem) ([]repository.StockAdjustmentItem, error) {
	merged := make(map[int64]float64, len(items))
	for _, it := range items {
		merged[it.ProductItemID] += float64(it.Quantity)

		comp, err := uc.componentRepo.GetByProductItem(ctx, it.ProductItemID, merchantID)
		if err != nil {
			if err == domain.ErrProductComponentNotFound {
				continue
			}
			return nil, err
		}
		for _, ci := range comp.Items {
			merged[ci.ComponentProductItemID] += ci.Quantity * ci.ConversionFactor * float64(it.Quantity)
		}
	}

	ids := make([]int64, 0, len(merged))
	for id := range merged {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	result := make([]repository.StockAdjustmentItem, 0, len(ids))
	for _, id := range ids {
		quantity := int64(math.Ceil(merged[id]))
		if quantity <= 0 {
			continue
		}
		result = append(result, repository.StockAdjustmentItem{ProductItemID: id, Quantity: quantity})
	}
	return result, nil
}

var _ StockUsecase = (*stockUsecase)(nil)
