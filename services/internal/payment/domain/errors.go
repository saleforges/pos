package domain

import "errors"

var (
	ErrPaymentNotFound        = errors.New("PAY001: payment not found")
	ErrInvalidPayment         = errors.New("PAY002: invalid payment data")
	ErrGatewayNotConfigured   = errors.New("PAY003: payment gateway not configured")
	ErrGatewayUnavailable     = errors.New("PAY004: payment gateway unavailable")
	ErrGatewayError           = errors.New("PAY005: payment gateway error")
	ErrInvalidCallback        = errors.New("PAY006: invalid payment callback")
	ErrDuplicateCallback      = errors.New("PAY007: duplicate payment callback")
	ErrOrderClientUnavailable = errors.New("PAY008: order service unavailable")
	ErrInternal               = errors.New("PAY500: internal error")
)
