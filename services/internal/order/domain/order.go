package domain

import "time"

type OrderStatus string

const (
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type PaymentStatus string

const (
	PaymentStatusUnpaid  PaymentStatus = "unpaid"
	PaymentStatusPartial PaymentStatus = "partial"
	PaymentStatusPaid    PaymentStatus = "paid"
)

type Order struct {
	ID            int64           `json:"id"`
	MerchantID    int64           `json:"merchantId"`
	BranchID      int64           `json:"branchId"`
	CreatedBy     int64           `json:"createdBy"`
	CustomerID    *int64          `json:"customerId,omitempty"`
	ClientOrderID string          `json:"clientOrderId,omitempty"`
	Status        OrderStatus     `json:"status"`
	PaymentStatus PaymentStatus   `json:"paymentStatus"`
	Subtotal      float64         `json:"subtotal"`
	Discount      float64         `json:"discount"`
	Tax           float64         `json:"tax"`
	Total         float64         `json:"total"`
	PaidAmount    float64         `json:"paidAmount"`
	DueDate       *time.Time      `json:"dueDate,omitempty"`
	Note          string          `json:"note,omitempty"`
	CreatedAt     time.Time       `json:"createdAt"`
	UpdatedAt     time.Time       `json:"updatedAt"`
	Items         []OrderItem     `json:"items,omitempty"`
	Payments      []PaymentRecord `json:"payments,omitempty"`
}

// ComputePaymentStatus derives the payment status from paid amount vs total.
func (o *Order) ComputePaymentStatus() PaymentStatus {
	switch {
	case o.PaidAmount >= o.Total:
		return PaymentStatusPaid
	case o.PaidAmount > 0:
		return PaymentStatusPartial
	default:
		return PaymentStatusUnpaid
	}
}

func (o *Order) Validate() error {
	if o.MerchantID == 0 {
		return ErrInvalidOrder
	}
	if o.BranchID == 0 {
		return ErrInvalidOrder
	}
	if len(o.Items) == 0 {
		return ErrInvalidOrder
	}
	for i := range o.Items {
		if err := o.Items[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}

type OrderItem struct {
	ID            int64   `json:"id"`
	OrderID       int64   `json:"orderId"`
	ProductItemID int64   `json:"productItemId"`
	ItemName      string  `json:"itemName"`
	UnitPrice     float64 `json:"unitPrice"`
	Quantity      float64 `json:"quantity"`
	LineTotal     float64 `json:"lineTotal"`
}

func (i *OrderItem) Validate() error {
	if i.ProductItemID == 0 {
		return ErrInvalidOrderItem
	}
	if i.ItemName == "" {
		return ErrInvalidOrderItem
	}
	if i.Quantity <= 0 {
		return ErrInvalidOrderItem
	}
	if i.UnitPrice < 0 {
		return ErrInvalidOrderItem
	}
	return nil
}

type PaymentMethod string

const (
	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodTransfer PaymentMethod = "transfer"
	PaymentMethodQRIS     PaymentMethod = "qris"
)

type PaymentRecord struct {
	ID        int64         `json:"id"`
	OrderID   int64         `json:"orderId"`
	Amount    float64       `json:"amount"`
	Method    PaymentMethod `json:"method"`
	CreatedBy int64         `json:"createdBy"`
	PaidAt    time.Time     `json:"paidAt"`
	CreatedAt time.Time     `json:"createdAt"`
}
