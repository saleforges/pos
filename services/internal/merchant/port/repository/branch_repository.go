package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/merchant/domain"
)

type BranchRepository interface {
	Create(ctx context.Context, branch *domain.Branch) error
	GetByID(ctx context.Context, id string) (*domain.Branch, error)
	ListByMerchant(ctx context.Context, merchantID string) ([]domain.Branch, error)
	Update(ctx context.Context, branch *domain.Branch) error
	Delete(ctx context.Context, id string) error
}
