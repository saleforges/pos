package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/catalog/domain"
	"github.com/saleforge/pos/services/pkg/pagination"
)

type ProductUsecase interface {
	Create(ctx context.Context, params CreateProductParams) (*domain.Product, error)
	GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Product, error)
	List(ctx context.Context, merchantID int64, search string, p pagination.Params) ([]domain.Product, *pagination.Metadata, error)
	Update(ctx context.Context, params UpdateProductParams) (*domain.Product, error)
	Delete(ctx context.Context, id int64, merchantID int64) error
}

type CreateProductParams struct {
	MerchantID  int64
	CategoryID  int64
	Name        string
	Description string
	ImageURL    string
}

type UpdateProductParams struct {
	ID          int64
	MerchantID  int64
	CategoryID  *int64
	Name        *string
	Description *string
	ImageURL    *string
	Status      *domain.ProductStatus
}
