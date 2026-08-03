package repository

import (
	"context"
)

// OrderClient notifies the order service when a payment succeeds, so the
// order can record the payment and advance its payment status.
type OrderClient interface {
	NotifyPaid(ctx context.Context, orderID, merchantID int64, amount float64, method string) error
}
