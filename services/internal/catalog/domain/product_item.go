package domain

import "time"

type ProductItemStatus string

const (
	ProductItemStatusActive   ProductItemStatus = "active"
	ProductItemStatusInactive ProductItemStatus = "inactive"
)

type ProductItem struct {
	ID             int64             `json:"id"`
	ProductID      int64             `json:"productId"`
	MerchantID     int64             `json:"merchantId"`
	Name           string            `json:"name"`
	SKU            string            `json:"sku,omitempty"`
	UnitID         *int64            `json:"unitId,omitempty"`
	Price          Price             `json:"price"`
	TrackInventory bool              `json:"trackInventory"`
	ImageURL       string            `json:"imageUrl,omitempty"`
	Status         ProductItemStatus `json:"status"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
	DeletedAt      *time.Time        `json:"deletedAt,omitempty"`
}

type Price struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type ProductItemBarcode struct {
	ID            int64  `json:"id"`
	ProductItemID int64  `json:"productItemId"`
	Barcode       string `json:"barcode"`
}
