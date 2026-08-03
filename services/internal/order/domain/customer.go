package domain

import "time"

type Customer struct {
	ID         int64     `json:"id"`
	MerchantID int64     `json:"merchantId"`
	Name       string    `json:"name"`
	Phone      string    `json:"phone,omitempty"`
	Address    string    `json:"address,omitempty"`
	Note       string    `json:"note,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (c *Customer) Validate() error {
	if c.MerchantID == 0 {
		return ErrInvalidCustomer
	}
	if c.Name == "" {
		return ErrInvalidCustomer
	}
	return nil
}
