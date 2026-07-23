package domain

import "errors"

var (
	ErrProductNotFound      = errors.New("CAT001: product not found")
	ErrCategoryNotFound     = errors.New("CAT002: category not found")
	ErrSellableItemNotFound = errors.New("CAT003: sellable item not found")
	ErrUnitNotFound         = errors.New("CAT004: unit not found")
	ErrBarcodeExists        = errors.New("CAT005: barcode already exists")
	ErrInvalidProduct       = errors.New("CAT006: invalid product data")
	ErrInvalidCategory      = errors.New("CAT007: invalid category data")
	ErrInvalidSellableItem  = errors.New("CAT008: invalid sellable item data")
	ErrInternal             = errors.New("CAT500: internal error")
)
