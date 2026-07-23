package product

import "github.com/saleforge/pos/services/internal/catalog/domain"

type createProductReq struct {
	CategoryID  int64   `json:"categoryId"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	ImageURL    string  `json:"imageUrl,omitempty"`
	Price       float64 `json:"price,omitempty"`
	Items       []bulkItemReq `json:"items,omitempty"`
}

type updateProductReq struct {
	CategoryID  *int64                `json:"categoryId,omitempty"`
	Name        *string               `json:"name,omitempty"`
	Description *string               `json:"description,omitempty"`
	ImageURL    *string               `json:"imageUrl,omitempty"`
	Status      *domain.ProductStatus `json:"status,omitempty"`
}

type createBulkReq struct {
	CategoryID  int64         `json:"categoryId"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	ImageURL    string        `json:"imageUrl,omitempty"`
	Items       []bulkItemReq `json:"items"`
}

type bulkItemReq struct {
	Name           string  `json:"name"`
	SKU            string  `json:"sku,omitempty"`
	UnitID         *int64  `json:"unitId,omitempty"`
	Price          float64 `json:"price"`
	TrackInventory bool    `json:"trackInventory"`
	ImageURL       string  `json:"imageUrl,omitempty"`
}

type updateBulkReq struct {
	CategoryID  *int64         `json:"categoryId,omitempty"`
	Name        *string        `json:"name,omitempty"`
	Description *string        `json:"description,omitempty"`
	ImageURL    *string        `json:"imageUrl,omitempty"`
	Status      *string        `json:"status,omitempty"`
	Items       *[]bulkItemReq `json:"items,omitempty"`
}
