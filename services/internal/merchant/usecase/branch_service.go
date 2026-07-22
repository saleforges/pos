package usecase

import (
	"context"

	"github.com/saleforge/pos/services/internal/merchant/domain"
	"github.com/saleforge/pos/services/pkg/pagination"
)

// BranchUsecase defines branch CRUD operations.
type BranchUsecase interface {
	CreateBranch(ctx context.Context, input CreateBranchParams) (*domain.Branch, error)
	GetBranch(ctx context.Context, id int64) (*domain.Branch, error)
	ListBranches(ctx context.Context, merchantID int64, p pagination.Params) ([]domain.Branch, *pagination.Metadata, error)
	UpdateBranch(ctx context.Context, input UpdateBranchParams) (*domain.Branch, error)
	DeleteBranch(ctx context.Context, id int64) error
}
