package domain

import "time"

type Category struct {
	ID         int64      `json:"id"`
	MerchantID int64      `json:"merchantId"`
	Name       string     `json:"name"`
	ParentID   *int64     `json:"parentId,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	DeletedAt  *time.Time `json:"deletedAt,omitempty"`
}
