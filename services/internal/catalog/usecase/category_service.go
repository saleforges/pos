package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/catalog/domain"
)

type CategoryUsecase interface {
	Create(ctx context.Context, params CreateCategoryParams) (*domain.Category, error)
	GetByID(ctx context.Context, id int64, merchantID int64) (*domain.Category, error)
	ListByMerchant(ctx context.Context, merchantID int64) ([]domain.Category, error)
	Update(ctx context.Context, params UpdateCategoryParams) (*domain.Category, error)
	Delete(ctx context.Context, id int64, merchantID int64) error
	Restore(ctx context.Context, id int64, merchantID int64) (*domain.Category, error)
}

type CreateCategoryParams struct {
	MerchantID int64
	Name       string
	ParentID   *int64
}

type UpdateCategoryParams struct {
	ID        int64
	MerchantID int64
	Name      *string
	ParentID  *int64
}
