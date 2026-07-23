package domain

import "time"

type SellableItemStatus string

const (
	SellableItemStatusActive   SellableItemStatus = "active"
	SellableItemStatusInactive SellableItemStatus = "inactive"
)

type SellableItem struct {
	ID             int64              `json:"id"`
	ProductID      int64              `json:"productId"`
	Name           string             `json:"name"`
	UnitID         int64              `json:"unitId"`
	Price          float64            `json:"price"`
	TrackInventory bool               `json:"trackInventory"`
	ImageURL       string             `json:"imageUrl,omitempty"`
	Status         SellableItemStatus `json:"status"`
	CreatedAt      time.Time          `json:"createdAt"`
	UpdatedAt      time.Time          `json:"updatedAt"`
	DeletedAt      *time.Time         `json:"deletedAt,omitempty"`
}

type SellableItemBarcode struct {
	ID             int64  `json:"id"`
	SellableItemID int64  `json:"sellableItemId"`
	Barcode        string `json:"barcode"`
}
