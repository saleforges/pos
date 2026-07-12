package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/merchant/domain"
)

type BranchRepository interface {
	Create(ctx context.Context, branch *domain.Branch) error
	GetByID(ctx context.Context, id int64) (*domain.Branch, error)
	ListByMerchant(ctx context.Context, merchantID int64) ([]domain.Branch, error)
	Update(ctx context.Context, branch *domain.Branch) error
	Delete(ctx context.Context, id int64) error
}
