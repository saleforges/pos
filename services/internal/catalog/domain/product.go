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
	MerchantID  int64         `json:"merchantId"`
	CategoryID  int64         `json:"categoryId,omitempty"`
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	ImageURL    string        `json:"imageUrl,omitempty"`
	Status      ProductStatus `json:"status"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
	DeletedAt   *time.Time    `json:"deletedAt,omitempty"`
}
