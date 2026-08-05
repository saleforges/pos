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
	unitRepo       repository.UnitRepository
}

func NewStockUsecase(repo repository.StockRepository, adjustmentRepo repository.StockAdjustmentRepository, componentRepo repository.ProductComponentRepository, unitRepo repository.UnitRepository) StockUsecase {
	return &stockUsecase{repo: repo, adjustmentRepo: adjustmentRepo, componentRepo: componentRepo, unitRepo: unitRepo}
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
// [EsTeh:2, Gula:2]. Each component quantity is normalized to the raw
// material's base stock unit via the unit's factor-to-base (e.g. 0.5 kg →
// 500 g) BEFORE merging, so items sharing a raw material in different units
// sum correctly. Fractional results are rounded up to never undersell.
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
			// Normalize component quantity to base stock unit. Unknown
			// units fall back to factor 1 (already in base unit).
			factor := 1.0
			if u, err := uc.unitRepo.GetByID(ctx, ci.UnitID); err == nil && u != nil {
				factor = u.FactorToBase
			}
			merged[ci.ComponentProductItemID] += ci.Quantity * factor * float64(it.Quantity)
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

func (uc *stockUsecase) Sync(ctx context.Context, merchantID, branchID int64, lastSync *time.Time) (*StockSyncResult, error) {
	stocks, err := uc.repo.SyncByBranch(ctx, merchantID, branchID, lastSync)
	if err != nil {
		return nil, err
	}
	if stocks == nil {
		stocks = []domain.Stock{}
	}
	// Server time becomes the next sync token.
	return &StockSyncResult{
		Stocks:    stocks,
		SyncToken: time.Now().UTC().Format(time.RFC3339Nano),
	}, nil
}

var _ StockUsecase = (*stockUsecase)(nil)
