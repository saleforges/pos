package repository

import (
	"context"

	"github.com/saleforge/pos/services/internal/merchant/domain"
)

type StaffRepository interface {
	Create(ctx context.Context, staff *domain.StaffMember) error
	GetByID(ctx context.Context, id int64) (*domain.StaffMember, error)
	ListByBranch(ctx context.Context, branchID int64) ([]domain.StaffMember, error)
	ListByMerchant(ctx context.Context, merchantID int64) ([]domain.StaffMember, error)
	GetByUserAndBranch(ctx context.Context, userID, branchID int64) (*domain.StaffMember, error)
	ListByUserAndMerchant(ctx context.Context, userID, merchantID int64) ([]domain.StaffMember, error)
	SetDefaultBranch(ctx context.Context, userID, branchID int64) error
	Update(ctx context.Context, staff *domain.StaffMember) error
	Delete(ctx context.Context, id int64) error
}
