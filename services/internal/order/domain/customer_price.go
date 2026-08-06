package domain

import "time"

type CustomerPrice struct {
	ID            int64     `json:"id"`
	MerchantID    int64     `json:"merchantId"`
	CustomerID    int64     `json:"customerId"`
	ProductItemID int64     `json:"productItemId"`
	Price         float64   `json:"price"`
	Currency      string    `json:"currency"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
