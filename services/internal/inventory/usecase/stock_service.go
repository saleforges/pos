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

func (uc *stockUsecase) Produce(ctx context.Context, params ProduceParams) (*domain.Stock, error) {
	if params.BranchID == 0 || params.ProductItemID == 0 || params.Quantity <= 0 {
		return nil, domain.ErrInvalidStock
	}

	expanded, err := uc.expandComponents(ctx, params.MerchantID, []AdjustStockItem{
		{ProductItemID: params.ProductItemID, Quantity: params.Quantity},
	})
	if err != nil {
		return nil, err
	}

	rawMaterials := make([]repository.StockAdjustmentItem, 0, len(expanded))
	for _, item := range expanded {
		if item.ProductItemID == params.ProductItemID {
			continue
		}
		rawMaterials = append(rawMaterials, item)
	}
	if len(rawMaterials) == 0 {
		return nil, domain.ErrProductComponentNotFound
	}

	if err := uc.adjustmentRepo.Deduct(ctx, params.MerchantID, params.BranchID, "production", 0, rawMaterials); err != nil {
		return nil, err
	}

	existing, err := uc.findStock(ctx, params.MerchantID, params.BranchID, params.ProductItemID)
	if err != nil {
		return nil, err
	}

	var produced *domain.Stock
	if existing == nil {
		produced, err = uc.Create(ctx, CreateStockParams{
			MerchantID:    params.MerchantID,
			BranchID:      params.BranchID,
			ProductItemID: params.ProductItemID,
			Available:     params.Quantity,
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := uc.adjustmentRepo.Restore(ctx, params.MerchantID, params.BranchID, "production", 0, []repository.StockAdjustmentItem{
			{ProductItemID: params.ProductItemID, Quantity: params.Quantity},
		}); err != nil {
			return nil, err
		}
		produced, err = uc.repo.GetByID(ctx, existing.ID, params.MerchantID)
		if err != nil {
			return nil, err
		}
	}

	if comp, err := uc.componentRepo.GetByProductItem(ctx, params.ProductItemID, params.MerchantID); err == nil {
		_ = uc.componentRepo.Delete(ctx, comp.ID, params.MerchantID)
	}

	return produced, nil
}

func (uc *stockUsecase) Opname(ctx context.Context, params OpnameParams) (*domain.Stock, error) {
	if params.BranchID == 0 || params.ProductItemID == 0 || params.ActualQuantity < 0 {
		return nil, domain.ErrInvalidStock
	}
	referenceType := params.Reason
	if referenceType == "" {
		referenceType = "opname"
	}

	existing, err := uc.findStock(ctx, params.MerchantID, params.BranchID, params.ProductItemID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return uc.Create(ctx, CreateStockParams{
			MerchantID:    params.MerchantID,
			BranchID:      params.BranchID,
			ProductItemID: params.ProductItemID,
			Available:     params.ActualQuantity,
		})
	}

	delta := params.ActualQuantity - existing.Available
	switch {
	case delta > 0:
		if err := uc.adjustmentRepo.Restore(ctx, params.MerchantID, params.BranchID, referenceType, 0, []repository.StockAdjustmentItem{
			{ProductItemID: params.ProductItemID, Quantity: delta},
		}); err != nil {
			return nil, err
		}
	case delta < 0:
		if err := uc.adjustmentRepo.Deduct(ctx, params.MerchantID, params.BranchID, referenceType, 0, []repository.StockAdjustmentItem{
			{ProductItemID: params.ProductItemID, Quantity: -delta},
		}); err != nil {
			return nil, err
		}
	}

	return uc.repo.GetByID(ctx, existing.ID, params.MerchantID)
}

func (uc *stockUsecase) findStock(ctx context.Context, merchantID, branchID, productItemID int64) (*domain.Stock, error) {
	stocks, err := uc.repo.List(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	for i := range stocks {
		if stocks[i].BranchID == branchID && stocks[i].ProductItemID == productItemID {
			return &stocks[i], nil
		}
	}
	return nil, nil
}

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

func (uc *stockUsecase) Transfer(ctx context.Context, params TransferStockParams) error {
	if params.FromBranchID == 0 || params.ToBranchID == 0 || len(params.Items) == 0 {
		return domain.ErrInvalidStock
	}
	items := make([]repository.StockAdjustmentItem, len(params.Items))
	for i, it := range params.Items {
		items[i] = repository.StockAdjustmentItem{ProductItemID: it.ProductItemID, Quantity: it.Quantity, ReferenceID: 0}
	}
	return uc.adjustmentRepo.Transfer(ctx, params.MerchantID, params.FromBranchID, params.ToBranchID, items)
}

func (uc *stockUsecase) Sync(ctx context.Context, merchantID, branchID int64, lastSync *time.Time) (*StockSyncResult, error) {
	stocks, err := uc.repo.SyncByBranch(ctx, merchantID, branchID, lastSync)
	if err != nil {
		return nil, err
	}
	if stocks == nil {
		stocks = []domain.Stock{}
	}
	return &StockSyncResult{
		Stocks:    stocks,
		SyncToken: time.Now().UTC().Format(time.RFC3339Nano),
	}, nil
}

var _ StockUsecase = (*stockUsecase)(nil)
