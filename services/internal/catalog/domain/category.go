package domain

import "time"

type CategoryStatus string

const (
	CategoryStatusActive   CategoryStatus = "active"
	CategoryStatusInactive CategoryStatus = "inactive"
)

type Category struct {
	ID          string         `json:"id"`
	MerchantID  string         `json:"merchant_id"`
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
	ParentID    *string        `json:"parent_id,omitempty"`
	SortOrder   int            `json:"sort_order"`
	Status      CategoryStatus `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
