package usecase

import (
	"context"
	"time"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type SyncUsecase interface {
	Sync(ctx context.Context, params SyncParams) (*SyncResult, error)
}

type SyncProductChange struct {
	ID          int64  `json:"id"`
	CategoryID  int64  `json:"categoryId"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ImageURL    string `json:"imageUrl,omitempty"`
	Status      string `json:"status,omitempty"`
}

type SyncItemChange struct {
	ID             int64   `json:"id"`
	ProductID      int64   `json:"productId"`
	Name           string  `json:"name"`
	SKU            string  `json:"sku,omitempty"`
	UnitID         *int64  `json:"unitId,omitempty"`
	PriceAmount    float64 `json:"price"`
	PriceCurrency  string  `json:"currency,omitempty"`
	TrackInventory bool    `json:"trackInventory"`
	ImageURL       string  `json:"imageUrl,omitempty"`
	Status         string  `json:"status,omitempty"`
}

type SyncCategoryChange struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	ParentID *int64 `json:"parentId,omitempty"`
}

type SyncParams struct {
	MerchantID int64
	BranchID   int64
	LastSync   *time.Time
	Products   []SyncProductChange
	Items      []SyncItemChange
	Categories []SyncCategoryChange
}

type SyncResult struct {
	SyncToken  string
	Products   []domain.Product
	Items      []domain.ProductItem
	Categories []domain.Category
	Units      []domain.Unit
	Barcodes   []domain.ProductItemBarcode
}
