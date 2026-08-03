package domain

import (
	"time"
)

type Stock struct {
	ID            int64     `json:"id"`
	MerchantID    int64     `json:"merchantId"`
	BranchID      int64     `json:"branchId"`
	ProductItemID int64     `json:"productItemId"`
	Available     int64     `json:"available"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// Validate checks stock field constraints.
func (s *Stock) Validate() error {
	if s.MerchantID == 0 {
		return ErrInvalidStock
	}
	if s.BranchID == 0 {
		return ErrInvalidStock
	}
	if s.ProductItemID == 0 {
		return ErrInvalidStock
	}
	if s.Available < 0 {
		return ErrNegativeAvailable
	}
	return nil
}
