package domain

import "errors"

var (
	ErrStockNotFound               = errors.New("INV001: stock not found")
	ErrInvalidStock                = errors.New("INV002: invalid stock data")
	ErrNegativeAvailable           = errors.New("INV003: available cannot be negative")
	ErrProductComponentNotFound    = errors.New("INV004: product component not found")
	ErrInvalidProductComponent     = errors.New("INV005: invalid product component data")
	ErrComponentAlreadyExists      = errors.New("INV006: product component already exists for this product item")
	ErrInvalidProductComponentItem = errors.New("INV007: invalid product component item data")
	ErrNoComponentItems            = errors.New("INV008: product component must contain at least one item")
	ErrStockMovementNotFound       = errors.New("INV009: stock movement not found")
	ErrInsufficientStock           = errors.New("INV010: insufficient stock")
	ErrInternal                    = errors.New("INV500: internal error")
)
