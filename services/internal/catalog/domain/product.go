package domain

import "time"

type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "active"
	ProductStatusInactive ProductStatus = "inactive"
	ProductStatusArchived ProductStatus = "archived"
)

type Product struct {
	ID          int64         `json:"id"`
	MerchantID  int64         `json:"merchant_id"`
	CategoryID  int64         `json:"category_id"`
	Name        string        `json:"name"`
	SKU         string        `json:"sku"`
	Barcode     string        `json:"barcode,omitempty"`
	Description string        `json:"description,omitempty"`
	Price       float64       `json:"price"`
	Cost        float64       `json:"cost,omitempty"`
	TaxRate     float64       `json:"tax_rate"`
	Unit        string        `json:"unit"`
	ImageURL    string        `json:"image_url,omitempty"`
	Status      ProductStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}
