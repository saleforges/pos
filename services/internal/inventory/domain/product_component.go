package domain

import (
	"time"
)

type ProductComponent struct {
	ID            int64     `json:"id"`
	MerchantID    int64     `json:"merchantId"`
	ProductItemID int64     `json:"productItemId"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	Items         []ProductComponentItem `json:"items,omitempty"`
}

func (pc *ProductComponent) Validate() error {
	if pc.MerchantID == 0 {
		return ErrInvalidProductComponent
	}
	if pc.ProductItemID == 0 {
		return ErrInvalidProductComponent
	}
	if len(pc.Items) == 0 {
		return ErrNoComponentItems
	}
	for i := range pc.Items {
		if err := pc.Items[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}

type ProductComponentItem struct {
	ID                   int64   `json:"id"`
	ProductComponentID   int64   `json:"productComponentId"`
	ComponentProductItemID int64 `json:"componentProductItemId"`
	Quantity             float64 `json:"quantity"`
	UnitID               int64   `json:"unitId"`
	CreatedAt            time.Time `json:"createdAt"`
}

func (pci *ProductComponentItem) Validate() error {
	if pci.ComponentProductItemID == 0 {
		return ErrInvalidProductComponentItem
	}
	if pci.Quantity <= 0 {
		return ErrInvalidProductComponentItem
	}
	if pci.UnitID == 0 {
		return ErrInvalidProductComponentItem
	}
	return nil
}
