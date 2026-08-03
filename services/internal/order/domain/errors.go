package domain

import "errors"

var (
	ErrOrderNotFound               = errors.New("ORD001: order not found")
	ErrInvalidOrder                = errors.New("ORD002: invalid order data")
	ErrInvalidOrderItem            = errors.New("ORD003: invalid order item")
	ErrInvalidTransition           = errors.New("ORD004: invalid status transition")
	ErrPaymentExceedsTotal         = errors.New("ORD005: payment exceeds remaining balance")
	ErrInvalidPayment              = errors.New("ORD006: invalid payment data")
	ErrCustomerNotFound            = errors.New("ORD007: customer not found")
	ErrInvalidCustomer             = errors.New("ORD008: invalid customer data")
	ErrInsufficientStock           = errors.New("ORD009: insufficient stock")
	ErrOrderStockNotFound          = errors.New("ORD010: stock not found for order item")
	ErrPaymentGatewayNotConfigured = errors.New("ORD011: payment gateway not configured")
	ErrPaymentGatewayUnavailable   = errors.New("ORD012: payment gateway unavailable")
	ErrPaymentGatewayError         = errors.New("ORD013: payment gateway error")
	ErrInvalidCallback             = errors.New("ORD014: invalid payment callback")
	ErrInternal                    = errors.New("ORD500: internal error")
)
