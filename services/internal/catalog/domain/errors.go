package domain

import "errors"

var (
	ErrProductNotFound    = errors.New("CAT001: product not found")
	ErrProductExists      = errors.New("product already exists")
	ErrCategoryNotFound   = errors.New("CAT002: category not found")
	ErrCategoryExists     = errors.New("category already exists")
	ErrCategoryHasProduct = errors.New("category has associated products")
	ErrVariantNotFound    = errors.New("CAT003: variant not found")
	ErrSkuExists          = errors.New("CAT004: sku already exists")
	ErrBarcodeExists      = errors.New("CAT005: barcode already exists")
	ErrInternal           = errors.New("internal server error")
)
