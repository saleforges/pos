package handler

import (
	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type categoryResponse struct {
	ID          int64                 `json:"id"`
	MerchantID  int64                 `json:"merchant_id"`
	Name        string                `json:"name"`
	Slug        string                `json:"slug"`
	Description string                `json:"description"`
	ParentID    *int64                `json:"parent_id,omitempty"`
	SortOrder   int                   `json:"sort_order"`
	Status      domain.CategoryStatus `json:"status"`
	CreatedAt   string                `json:"created_at"`
	UpdatedAt   string                `json:"updated_at"`
}

type productResponse struct {
	ID          int64                `json:"id"`
	MerchantID  int64                `json:"merchant_id"`
	CategoryID  int64                `json:"category_id"`
	Name        string               `json:"name"`
	SKU         string               `json:"sku"`
	Barcode     string               `json:"barcode,omitempty"`
	Description string               `json:"description,omitempty"`
	Price       float64              `json:"price"`
	Cost        float64              `json:"cost,omitempty"`
	TaxRate     float64              `json:"tax_rate"`
	Unit        string               `json:"unit"`
	ImageURL    string               `json:"image_url,omitempty"`
	Status      domain.ProductStatus `json:"status"`
	CreatedAt   string               `json:"created_at"`
	UpdatedAt   string               `json:"updated_at"`
}

type variantResponse struct {
	ID        int64   `json:"id"`
	ProductID int64   `json:"product_id"`
	Name      string  `json:"name"`
	SKU       string  `json:"sku"`
	Barcode   string  `json:"barcode,omitempty"`
	Price     float64 `json:"price"`
	Cost      float64 `json:"cost,omitempty"`
	ImageURL  string  `json:"image_url,omitempty"`
	SortOrder int     `json:"sort_order"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type createCategoryReq struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	ParentID    *int64 `json:"parent_id,omitempty"`
	SortOrder   int    `json:"sort_order"`
}

type updateCategoryReq struct {
	Name        *string                `json:"name,omitempty"`
	Slug        *string                `json:"slug,omitempty"`
	Description *string                `json:"description,omitempty"`
	ParentID    *int64                 `json:"parent_id,omitempty"`
	SortOrder   *int                   `json:"sort_order,omitempty"`
	Status      *domain.CategoryStatus `json:"status,omitempty"`
}

type createProductReq struct {
	CategoryID  int64   `json:"category_id"`
	Name        string  `json:"name"`
	SKU         string  `json:"sku"`
	Barcode     string  `json:"barcode,omitempty"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price"`
	Cost        float64 `json:"cost,omitempty"`
	TaxRate     float64 `json:"tax_rate"`
	Unit        string  `json:"unit"`
	ImageURL    string  `json:"image_url,omitempty"`
}

type updateProductReq struct {
	CategoryID  *int64                `json:"category_id,omitempty"`
	Name        *string               `json:"name,omitempty"`
	SKU         *string               `json:"sku,omitempty"`
	Barcode     *string               `json:"barcode,omitempty"`
	Description *string               `json:"description,omitempty"`
	Price       *float64              `json:"price,omitempty"`
	Cost        *float64              `json:"cost,omitempty"`
	TaxRate     *float64              `json:"tax_rate,omitempty"`
	Unit        *string               `json:"unit,omitempty"`
	ImageURL    *string               `json:"image_url,omitempty"`
	Status      *domain.ProductStatus `json:"status,omitempty"`
}

type createVariantReq struct {
	Name      string  `json:"name"`
	SKU       string  `json:"sku"`
	Barcode   string  `json:"barcode,omitempty"`
	Price     float64 `json:"price"`
	Cost      float64 `json:"cost,omitempty"`
	ImageURL  string  `json:"image_url,omitempty"`
	SortOrder int     `json:"sort_order"`
}

type updateVariantReq struct {
	Name      *string  `json:"name,omitempty"`
	SKU       *string  `json:"sku,omitempty"`
	Barcode   *string  `json:"barcode,omitempty"`
	Price     *float64 `json:"price,omitempty"`
	Cost      *float64 `json:"cost,omitempty"`
	ImageURL  *string  `json:"image_url,omitempty"`
	SortOrder *int     `json:"sort_order,omitempty"`
}

type paginatedMeta struct {
	Total  int `json:"total"`
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type paginatedResponse[T any] struct {
	Items []T           `json:"items"`
	Meta  paginatedMeta `json:"meta"`
}

func toCategoryResponse(c domain.Category) categoryResponse {
	return categoryResponse{
		ID:          c.ID,
		MerchantID:  c.MerchantID,
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		ParentID:    c.ParentID,
		SortOrder:   c.SortOrder,
		Status:      c.Status,
		CreatedAt:   c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toProductResponse(p domain.Product) productResponse {
	return productResponse{
		ID:          p.ID,
		MerchantID:  p.MerchantID,
		CategoryID:  p.CategoryID,
		Name:        p.Name,
		SKU:         p.SKU,
		Barcode:     p.Barcode,
		Description: p.Description,
		Price:       p.Price,
		Cost:        p.Cost,
		TaxRate:     p.TaxRate,
		Unit:        p.Unit,
		ImageURL:    p.ImageURL,
		Status:      p.Status,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toVariantResponse(v domain.Variant) variantResponse {
	return variantResponse{
		ID:        v.ID,
		ProductID: v.ProductID,
		Name:      v.Name,
		SKU:       v.SKU,
		Barcode:   v.Barcode,
		Price:     v.Price,
		Cost:      v.Cost,
		ImageURL:  v.ImageURL,
		SortOrder: v.SortOrder,
		CreatedAt: v.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: v.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
