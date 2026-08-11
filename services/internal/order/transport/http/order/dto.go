package order

import (
	"time"
)

type createOrderReq struct {
	BranchID      int64                `json:"branchId"`
	CustomerID    *int64               `json:"customerId,omitempty"`
	ClientOrderID string               `json:"clientOrderId,omitempty"`
	DueDate       *string              `json:"dueDate,omitempty"`
	Note          string               `json:"note,omitempty"`
	Discount      float64              `json:"discount,omitempty"`
	Items         []createOrderItemReq `json:"items"`
}

type createOrderItemReq struct {
	ProductItemID int64   `json:"productItemId"`
	ItemName      string  `json:"itemName"`
	UnitPrice     float64 `json:"unitPrice"`
	Quantity      float64 `json:"quantity"`
}

type cancelOrderReq struct {
	Status string `json:"status"`
}

type addPaymentReq struct {
	Amount float64 `json:"amount"`
	Method string  `json:"method"`
	PaidAt *string `json:"paidAt,omitempty"`
}

type updateOrderReq struct {
	DueDate *string `json:"dueDate,omitempty"`
	Note    *string `json:"note,omitempty"`
}

func parseDueDate(s *string) (*time.Time, error) {
	if s == nil {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", *s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
