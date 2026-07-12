package domain

import "time"

type Variant struct {
	ID        int64     `json:"id"`
	ProductID int64     `json:"product_id"`
	Name      string    `json:"name"`
	SKU       string    `json:"sku"`
	Barcode   string    `json:"barcode,omitempty"`
	Price     float64   `json:"price"`
	Cost      float64   `json:"cost,omitempty"`
	ImageURL  string    `json:"image_url,omitempty"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
