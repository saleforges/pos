package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/port/repository"
	"github.com/saleforge/pos/services/pkg/otel"
)

type syncUsecase struct {
	prodRepo    repository.ProductRepository
	itemRepo    repository.ProductItemRepository
	catRepo     repository.CategoryRepository
	unitRepo    repository.UnitRepository
	barcodeRepo repository.ProductItemBarcodeRepository
}

func NewSyncUsecase(
	prodRepo repository.ProductRepository,
	itemRepo repository.ProductItemRepository,
	catRepo repository.CategoryRepository,
	unitRepo repository.UnitRepository,
	barcodeRepo repository.ProductItemBarcodeRepository,
) SyncUsecase {
	return &syncUsecase{
		prodRepo:    prodRepo,
		itemRepo:    itemRepo,
		catRepo:     catRepo,
		unitRepo:    unitRepo,
		barcodeRepo: barcodeRepo,
	}
}

func (uc *syncUsecase) Sync(ctx context.Context, params SyncParams) (*SyncResult, error) {
	ctx, span := otel.StartSpan(ctx, "sync.Sync")
	defer span.End()

	merchantID := params.MerchantID

	// Step 1: Apply incoming changes from mobile (push)
	if err := uc.applyChanges(ctx, merchantID, params); err != nil {
		return nil, err
	}

	// Step 2: Determine sync time window
	since := time.Time{}
	if params.LastSync != nil {
		since = *params.LastSync
	}

	// Step 3: Pull changes since lastSync
	products, err := uc.prodRepo.ListUpdatedAfter(ctx, merchantID, since)
	if err != nil {
		return nil, err
	}

	items, err := uc.itemRepo.ListUpdatedAfter(ctx, merchantID, since)
	if err != nil {
		return nil, err
	}

	categories, err := uc.catRepo.ListUpdatedAfter(ctx, merchantID, since)
	if err != nil {
		return nil, err
	}

	units, err := uc.unitRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	barcodes, err := uc.barcodeRepo.ListByMerchant(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	// Step 4: Return sync token (server's current time)
	syncToken := time.Now().UTC().Format(time.RFC3339Nano)

	return &SyncResult{
		SyncToken:  syncToken,
		Products:   products,
		Items:      items,
		Categories: categories,
		Units:      units,
		Barcodes:   barcodes,
	}, nil
}

func (uc *syncUsecase) applyChanges(ctx context.Context, merchantID int64, params SyncParams) error {
	now := time.Now().UTC()

	for _, p := range params.Products {
		if p.ID == 0 {
			// Create new product
			product := &domain.Product{
				MerchantID:  merchantID,
				CategoryID:  p.CategoryID,
				Name:        p.Name,
				Description: p.Description,
				ImageURL:    p.ImageURL,
				Status:      domain.ProductStatus(domainOrDefault(p.Status, string(domain.ProductStatusActive))),
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			// Ignore per-item errors to allow partial sync
			_ = uc.prodRepo.Create(ctx, product)
		} else {
			// Update existing product
			existing, err := uc.prodRepo.GetByID(ctx, p.ID, merchantID)
			if err != nil {
				continue
			}
			existing.CategoryID = p.CategoryID
			existing.Name = p.Name
			existing.Description = p.Description
			existing.ImageURL = p.ImageURL
			if p.Status != "" {
				existing.Status = domain.ProductStatus(p.Status)
			}
			existing.UpdatedAt = now
			_ = uc.prodRepo.Update(ctx, existing)
		}
	}

	for _, item := range params.Items {
		currency := item.PriceCurrency
		if currency == "" {
			currency = "IDR"
		}
		if item.ID == 0 {
			// Create new item
			it := &domain.ProductItem{
				ProductID:      item.ProductID,
				MerchantID:     merchantID,
				Name:           item.Name,
				SKU:            item.SKU,
				UnitID:         item.UnitID,
				Price:          domain.Price{Amount: item.PriceAmount, Currency: currency},
				TrackInventory: item.TrackInventory,
				ImageURL:       item.ImageURL,
				Status:         domain.ProductItemStatus(domainOrDefault(item.Status, string(domain.ProductItemStatusActive))),
				CreatedAt:      now,
				UpdatedAt:      now,
			}
			_ = uc.itemRepo.Create(ctx, it)
		} else {
			// Update existing item
			existing, err := uc.itemRepo.GetByID(ctx, item.ID, merchantID)
			if err != nil {
				continue
			}
			existing.Name = item.Name
			existing.SKU = item.SKU
			existing.UnitID = item.UnitID
			existing.Price = domain.Price{Amount: item.PriceAmount, Currency: currency}
			existing.TrackInventory = item.TrackInventory
			existing.ImageURL = item.ImageURL
			if item.Status != "" {
				existing.Status = domain.ProductItemStatus(item.Status)
			}
			existing.UpdatedAt = now
			_ = uc.itemRepo.Update(ctx, existing)
		}
	}

	for _, c := range params.Categories {
		if c.ID == 0 {
			cat := &domain.Category{
				MerchantID: merchantID,
				Name:       c.Name,
				ParentID:   c.ParentID,
				CreatedAt:  now,
				UpdatedAt:  now,
			}
			_ = uc.catRepo.Create(ctx, cat)
		} else {
			existing, err := uc.catRepo.GetByID(ctx, c.ID, merchantID)
			if err != nil {
				continue
			}
			existing.Name = c.Name
			existing.ParentID = c.ParentID
			existing.UpdatedAt = now
			_ = uc.catRepo.Update(ctx, existing)
		}
	}

	return nil
}

func domainOrDefault(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
