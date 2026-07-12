package domain

import "time"

type BranchStatus string

const (
	BranchStatusActive   BranchStatus = "active"
	BranchStatusInactive BranchStatus = "inactive"
	BranchStatusClosed   BranchStatus = "closed"
)

type OperatingHours struct {
	Open  string `json:"open"`
	Close string `json:"close"`
}

type Branch struct {
	ID             int64            `json:"id"`
	MerchantID     int64            `json:"merchant_id"`
	Name           string          `json:"name"`
	Code           string          `json:"code"`
	Address        string          `json:"address"`
	Phone          string          `json:"phone"`
	Status         BranchStatus    `json:"status"`
	OperatingDays  []string        `json:"operating_days,omitempty"`
	OperatingHours *OperatingHours `json:"operating_hours,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}
