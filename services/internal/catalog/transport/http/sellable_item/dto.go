package sellableitem

import "github.com/saleforge/pos/services/internal/catalog/domain"

type createSellableItemReq struct {
	Name           string  `json:"name"`
	UnitID         int64   `json:"unitId"`
	Price          float64 `json:"price"`
	TrackInventory bool    `json:"trackInventory"`
	ImageURL       string  `json:"imageUrl,omitempty"`
}

type updateSellableItemReq struct {
	Name           *string                     `json:"name,omitempty"`
	UnitID         *int64                      `json:"unitId,omitempty"`
	Price          *float64                    `json:"price,omitempty"`
	TrackInventory *bool                       `json:"trackInventory,omitempty"`
	ImageURL       *string                     `json:"imageUrl,omitempty"`
	Status         *domain.SellableItemStatus  `json:"status,omitempty"`
}
