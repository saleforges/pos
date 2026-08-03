package domain

import "errors"

var (
	ErrOrderNotFound       = errors.New("ORD001: order not found")
	ErrInvalidOrder        = errors.New("ORD002: invalid order data")
	ErrInvalidOrderItem    = errors.New("ORD003: invalid order item")
	ErrInvalidTransition   = errors.New("ORD004: invalid status transition")
	ErrPaymentExceedsTotal = errors.New("ORD005: payment exceeds remaining balance")
	ErrInvalidPayment      = errors.New("ORD006: invalid payment data")
	ErrCustomerNotFound    = errors.New("ORD007: customer not found")
	ErrInvalidCustomer     = errors.New("ORD008: invalid customer data")
	ErrInsufficientStock   = errors.New("ORD009: insufficient stock")
	ErrOrderStockNotFound  = errors.New("ORD010: stock not found for order item")
	ErrInternal            = errors.New("ORD500: internal error")
)
