package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type CategoryUsecase interface {
	Create(ctx context.Context, input CreateCategoryInput) (*domain.Category, error)
	GetByID(ctx context.Context, id int64) (*domain.Category, error)
	List(ctx context.Context, merchantID int64, search string, offset, limit int) (*PaginatedResult[domain.Category], error)
	Update(ctx context.Context, input UpdateCategoryInput) (*domain.Category, error)
	Delete(ctx context.Context, id int64) error
}

type ProductUsecase interface {
	Create(ctx context.Context, input CreateProductInput) (*domain.Product, error)
	GetByID(ctx context.Context, id int64) (*domain.Product, error)
	List(ctx context.Context, merchantID int64, search string, offset, limit int) (*PaginatedResult[domain.Product], error)
	Update(ctx context.Context, input UpdateProductInput) (*domain.Product, error)
	Delete(ctx context.Context, id int64) error
}

type VariantUsecase interface {
	Create(ctx context.Context, input CreateVariantInput) (*domain.Variant, error)
	ListByProduct(ctx context.Context, productID int64) ([]domain.Variant, error)
	Update(ctx context.Context, input UpdateVariantInput) (*domain.Variant, error)
	Delete(ctx context.Context, id int64) error
}

type PaginatedResult[T any] struct {
	Items []T
	Meta  PaginationMeta
}

type PaginationMeta struct {
	Total  int
	Offset int
	Limit  int
}

type CreateCategoryInput struct {
	MerchantID  int64
	Name        string
	Slug        string
	Description string
	ParentID    *int64
	SortOrder   int
}

type UpdateCategoryInput struct {
	ID          int64
	Name        *string
	Slug        *string
	Description *string
	ParentID    *int64
	SortOrder   *int
	Status      *domain.CategoryStatus
}

type CreateProductInput struct {
	MerchantID  int64
	CategoryID  int64
	Name        string
	SKU         string
	Barcode     string
	Description string
	Price       float64
	Cost        float64
	TaxRate     float64
	Unit        string
	ImageURL    string
}

type UpdateProductInput struct {
	ID          int64
	CategoryID  *int64
	Name        *string
	SKU         *string
	Barcode     *string
	Description *string
	Price       *float64
	Cost        *float64
	TaxRate     *float64
	Unit        *string
	ImageURL    *string
	Status      *domain.ProductStatus
}

type CreateVariantInput struct {
	ProductID int64
	Name      string
	SKU       string
	Barcode   string
	Price     float64
	Cost      float64
	ImageURL  string
	SortOrder int
}

type UpdateVariantInput struct {
	ID        int64
	Name      *string
	SKU       *string
	Barcode   *string
	Price     *float64
	Cost      *float64
	ImageURL  *string
	SortOrder *int
}
