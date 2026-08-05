package sync

import (
	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/internal/catalog/usecase"
)

// Request DTO
type syncReq struct {
	LastSync   *string                      `json:"lastSync,omitempty"`
	Products   []usecase.SyncProductChange  `json:"products,omitempty"`
	Items      []usecase.SyncItemChange     `json:"items,omitempty"`
	Categories []usecase.SyncCategoryChange `json:"categories,omitempty"`
}

// Response DTO
type syncResp struct {
	SyncToken  string                      `json:"syncToken"`
	Products   []domain.Product            `json:"products"`
	Items      []domain.ProductItem        `json:"items"`
	Categories []domain.Category           `json:"categories"`
	Units      []domain.Unit               `json:"units"`
	Barcodes   []domain.ProductItemBarcode `json:"barcodes"`
}
