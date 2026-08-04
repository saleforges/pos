package domain

import "time"

type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusExpired PaymentStatus = "expired"
	PaymentStatusFailed  PaymentStatus = "failed"
)

// PaymentTransaction tracks one gateway payment attempt for an order.
type PaymentTransaction struct {
	ID         int64         `json:"id"`
	MerchantID int64         `json:"merchantId"`
	OrderID    int64         `json:"orderId"`
	Gateway    string        `json:"gateway"`
	Status     PaymentStatus `json:"status"`
	Amount     float64       `json:"amount"`
	PaymentURL string        `json:"paymentUrl,omitempty"`
	PaymentNo  string        `json:"paymentNo,omitempty"`
	QrString   string        `json:"qrString,omitempty"`
	QrImage    string        `json:"qrImage,omitempty"`
	ExpiredAt  string        `json:"expiredAt,omitempty"`
	SessionID  string        `json:"sessionId,omitempty"`
	GatewayRef string        `json:"gatewayRef,omitempty"`
	CreatedAt  time.Time     `json:"createdAt"`
	UpdatedAt  time.Time     `json:"updatedAt"`
}

func (p *PaymentTransaction) Validate() error {
	if p.MerchantID == 0 {
		return ErrInvalidPayment
	}
	if p.OrderID == 0 {
		return ErrInvalidPayment
	}
	if p.Gateway == "" {
		return ErrInvalidPayment
	}
	if p.Amount <= 0 {
		return ErrInvalidPayment
	}
	return nil
}
