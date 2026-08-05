package productitem

type createProductItemReq struct {
	Name           string  `json:"name"`
	SKU            string  `json:"sku,omitempty"`
	UnitID         *int    `json:"unitId,omitempty"`
	Price          float64 `json:"price"`
	Currency       string  `json:"currency,omitempty"`
	TrackInventory bool    `json:"trackInventory"`
	ImageURL       string  `json:"imageUrl,omitempty"`
}

type updateProductItemReq struct {
	Name           *string  `json:"name,omitempty"`
	SKU            *string  `json:"sku,omitempty"`
	UnitID         *int     `json:"unitId,omitempty"`
	Price          *float64 `json:"price,omitempty"`
	Currency       *string  `json:"currency,omitempty"`
	TrackInventory *bool    `json:"trackInventory,omitempty"`
	ImageURL       *string  `json:"imageUrl,omitempty"`
	Status         *string  `json:"status,omitempty"`
}
